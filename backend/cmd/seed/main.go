package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/db"
)

// Demo constants
const (
	demoEmail    = "demo@spendle.app"
	demoPassword = "demo12345"
	demoName     = "Demo User"
)

func main() {
	// Go Concept - Structured Logging:
	// slog provides structured key-value output for clear CLI execution feedback.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 1. Load configuration (uses existing .env and environment variable loading)
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 2. Connect to PostgreSQL database pool
	logger.Info("connecting to PostgreSQL database...")
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	// Go Concept - defer:
	// Guarantees pool connections are cleanly terminated when the main function finishes.
	defer pool.Close()

	// 3. Ensure database schema is migrated to the latest version
	logger.Info("ensuring migrations are up to date...")
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		logger.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	// 4. Idempotent cleanup: Delete existing demo user and cascaded data
	// 4. Idempotent cleanup: Delete existing demo user and cascaded data
	logger.Info("cleaning up existing demo user (if present)...")

	cleanupTx, err := pool.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin cleanup transaction", "error", err)
		os.Exit(1)
	}

	cleanupQueries := []string{
		`DELETE FROM goal_contributions WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
		`DELETE FROM goals WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
		`DELETE FROM month_closings WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
		`DELETE FROM transactions WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
		`DELETE FROM monthly_budgets WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
		`DELETE FROM categories WHERE user_id IN (SELECT id FROM users WHERE email = $1)`,
		`DELETE FROM users WHERE email = $1`,
	}

	for _, query := range cleanupQueries {
		if _, err := cleanupTx.Exec(ctx, query, demoEmail); err != nil {
			_ = cleanupTx.Rollback(ctx)
			logger.Error("failed to clean up demo user", "error", err)
			os.Exit(1)
		}
	}

	if err := cleanupTx.Commit(ctx); err != nil {
		logger.Error("failed to commit cleanup transaction", "error", err)
		os.Exit(1)
	}

	// 5. Create Demo User with bcrypt hashed password (cost 12)
	logger.Info("creating demo user account...", "email", demoEmail)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), 12)
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		os.Exit(1)
	}

	var userID string
	userInsertQuery := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	if err := pool.QueryRow(ctx, userInsertQuery, demoName, demoEmail, string(passwordHash)).Scan(&userID); err != nil {
		logger.Error("failed to insert demo user", "error", err)
		os.Exit(1)
	}

	// 6. Create default categories + custom "Online Food" category
	categories := []string{
		"Food", "Groceries", "Transport", "Bills",
		"Shopping", "Entertainment", "Health", "Other",
		"Online Food",
	}

	categoryIDs := make(map[string]string)
	catInsertQuery := `
		INSERT INTO categories (user_id, name)
		VALUES ($1, $2)
		RETURNING id;
	`
	for _, catName := range categories {
		var catID string
		if err := pool.QueryRow(ctx, catInsertQuery, userID, catName).Scan(&catID); err != nil {
			logger.Error("failed to insert category", "name", catName, "error", err)
			os.Exit(1)
		}
		categoryIDs[catName] = catID
	}
	logger.Info("created categories", "count", len(categoryIDs))

	// 7. Calculate Relative Months (M0 = Current, M1 = Previous, M2 = 2 Months Ago)
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	m0 := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	m1 := m0.AddDate(0, -1, 0)
	m2 := m0.AddDate(0, -2, 0)

	// 8. Seed Monthly Budgets for M0, M1, and M2
	// Budgets in INR: Food 6000, Groceries 5000, Transport 2500, Bills 4000, Online Food 1500, Entertainment 2000
	// Strict Monetary Rule: All amounts stored as integer paise (1 Rupee = 100 Paise).
	budgetConfig := map[string]int64{
		"Food":          600000, // ₹6,000
		"Groceries":     500000, // ₹5,000
		"Transport":     250000, // ₹2,500
		"Bills":         400000, // ₹4,000
		"Online Food":   150000, // ₹1,500
		"Entertainment": 200000, // ₹2,000
	}

	budgetInsertQuery := `
		INSERT INTO monthly_budgets (user_id, category_id, month, limit_paise)
		VALUES ($1, $2, $3, $4);
	`
	for _, m := range []time.Time{m2, m1, m0} {
		for catName, limitPaise := range budgetConfig {
			catID := categoryIDs[catName]
			if _, err := pool.Exec(ctx, budgetInsertQuery, userID, catID, m, limitPaise); err != nil {
				logger.Error("failed to insert budget", "month", m.Format("2006-01"), "category", catName, "error", err)
				os.Exit(1)
			}
		}
	}
	logger.Info("seeded budgets for M2, M1, and M0")

	// 9. Seed Income Transactions
	// Salary ₹45,000 on the 1st of M2, M1, and M0; and Freelance ₹8,000 in M1
	incomeQuery := `
		INSERT INTO transactions (user_id, type, amount_paise, category_id, payment_mode, note, occurred_on)
		VALUES ($1, 'income', $2, NULL, $3, $4, $5);
	`
	incomes := []struct {
		amountPaise int64
		mode        string
		note        string
		date        time.Time
	}{
		{4500000, "netbanking", "Monthly Tech Salary", m2},
		{4500000, "netbanking", "Monthly Tech Salary", m1},
		{800000, "upi", "Freelance React Consulting", m1.AddDate(0, 0, 14)},
		{4500000, "netbanking", "Monthly Tech Salary", m0},
	}

	for _, inc := range incomes {
		if _, err := pool.Exec(ctx, incomeQuery, userID, inc.amountPaise, inc.mode, inc.note, inc.date); err != nil {
			logger.Error("failed to insert income", "note", inc.note, "error", err)
			os.Exit(1)
		}
	}
	logger.Info("seeded income transactions")

	// 10. Seed Expenses
	txInsertQuery := `
		INSERT INTO transactions (user_id, type, amount_paise, category_id, payment_mode, note, occurred_on)
		VALUES ($1, 'expense', $2, $3, $4, $5, $6);
	`

	type expenseItem struct {
		category string
		amount   int64 // paise
		mode     string
		note     string
		day      int
	}

	// M2 Expenses (Total budgeted = ₹21,000; total spent = ₹12,947; unspent = ₹8,053, within ₹6,000-9,000)
	m2Expenses := []expenseItem{
		{"Food", 140000, "upi", "D-Mart weekend snacks", 3},
		{"Groceries", 280000, "card", "BigBasket monthly pantry staples", 5},
		{"Transport", 80000, "upi", "Metro card recharge", 6},
		{"Bills", 119900, "upi", "Airtel Xstream broadband bill", 7},
		{"Online Food", 45000, "upi", "Zomato Meghana Foods biryani", 9},
		{"Food", 35000, "cash", "Team lunch sandwich", 12},
		{"Entertainment", 49900, "card", "Netflix standard subscription", 14},
		{"Transport", 100000, "card", "Shell petrol refill", 16},
		{"Groceries", 65000, "cash", "Fresh vegetables & fruits", 18},
		{"Bills", 39900, "upi", "Jio monthly prepaid recharge", 20},
		{"Online Food", 55000, "upi", "Swiggy dinner bowl", 22},
		{"Entertainment", 80000, "upi", "BookMyShow cinema tickets", 24},
		{"Food", 150000, "upi", "Dinner with college friends", 26},
		{"Transport", 40000, "upi", "Uber auto ride", 27},
		{"Groceries", 95000, "upi", "Organic dairy & eggs", 28},
	}

	for _, exp := range m2Expenses {
		occurred := m2.AddDate(0, 0, exp.day-1)
		catID := categoryIDs[exp.category]
		if _, err := pool.Exec(ctx, txInsertQuery, userID, exp.amount, catID, exp.mode, exp.note, occurred); err != nil {
			logger.Error("failed to insert M2 expense", "note", exp.note, "error", err)
			os.Exit(1)
		}
	}
	logger.Info("seeded M2 expenses", "count", len(m2Expenses))

	// M1 Expenses:
	// Total budgeted = ₹21,000; total spent = ₹13,498; unspent = ₹7,502 (within ₹6,000-9,000).
	// Online Food limit = ₹1,500; total spent = ₹1,950 (OVERSPENT by ₹450 to showcase coral progress bar).
	m1Expenses := []expenseItem{
		{"Food", 30000, "cash", "Irani Chai & Bun Maska", 2},
		{"Online Food", 55000, "upi", "Zomato Friday pizza order", 4},
		{"Groceries", 260000, "card", "BigBasket grocery delivery", 5},
		{"Bills", 145000, "netbanking", "BESCOM electricity bill", 8},
		{"Online Food", 65000, "upi", "Swiggy gourmet dinner", 11},
		{"Transport", 120000, "upi", "Petrol full tank refill", 12},
		{"Food", 180000, "card", "Weekend dining at Toit", 14},
		{"Entertainment", 29900, "upi", "Amazon Prime subscription", 16},
		{"Online Food", 45000, "upi", "Swiggy Instamart midnight snacks", 18},
		{"Transport", 60000, "upi", "Metro smart card recharge", 20},
		{"Bills", 39900, "upi", "Jio recharge for phone", 22},
		{"Groceries", 110000, "upi", "Local supermarket groceries", 24},
		{"Online Food", 30000, "upi", "Late night ice cream & waffles", 25},
		{"Entertainment", 80000, "upi", "Amoeba bowling with friends", 27},
		{"Transport", 90000, "cash", "Auto rickshaw commute", 28},
	}

	for _, exp := range m1Expenses {
		occurred := m1.AddDate(0, 0, exp.day-1)
		catID := categoryIDs[exp.category]
		if _, err := pool.Exec(ctx, txInsertQuery, userID, exp.amount, catID, exp.mode, exp.note, occurred); err != nil {
			logger.Error("failed to insert M1 expense", "note", exp.note, "error", err)
			os.Exit(1)
		}
	}
	logger.Info("seeded M1 expenses with overspent Online Food", "count", len(m1Expenses))

	// M0 Expenses (Current Month):
	// 5 transactions, strictly on or before today (never in the future)
	maxDay := today.Day()
	if maxDay < 1 {
		maxDay = 1
	}

	type m0Template struct {
		category string
		amount   int64
		mode     string
		note     string
		dayRatio float64 // used to distribute dates up to today
	}

	m0Templates := []m0Template{
		{"Groceries", 125000, "upi", "Weekly grocery replenishment", 0.15},
		{"Transport", 35000, "upi", "Metro commute", 0.35},
		{"Food", 42000, "upi", "Lunch with colleagues", 0.55},
		{"Online Food", 38000, "upi", "Swiggy quick lunch bowl", 0.75},
		{"Bills", 65000, "netbanking", "Water utility bill", 0.95},
	}

	for _, tmpl := range m0Templates {
		targetDay := int(float64(maxDay)*tmpl.dayRatio) + 1
		if targetDay > maxDay {
			targetDay = maxDay
		}
		occurred := m0.AddDate(0, 0, targetDay-1)
		catID := categoryIDs[tmpl.category]
		if _, err := pool.Exec(ctx, txInsertQuery, userID, tmpl.amount, catID, tmpl.mode, tmpl.note, occurred); err != nil {
			logger.Error("failed to insert M0 expense", "note", tmpl.note, "error", err)
			os.Exit(1)
		}
	}
	logger.Info("seeded M0 expenses strictly on or before today")

	// 11. Seed Savings Goals (saved_paise = 0 so month-end rollover effect is immediately visible)
	// - "Bike" ₹1,50,000 with deadline 6 months from now, priority 1
	// - "MacBook Air" ₹90,000 with monthly amount ₹7,500, priority 2
	// - "Goa trip" ₹25,000 with deadline 3 months from now, priority 3
	goalInsertQuery := `
		INSERT INTO goals (user_id, title, target_paise, saved_paise, deadline, monthly_amount_paise, priority, status)
		VALUES ($1, $2, $3, 0, $4, $5, $6, 'active');
	`

	bikeDeadline := m0.AddDate(0, 6, 0)
	goaDeadline := m0.AddDate(0, 3, 0)
	macbookMonthly := int64(750000) // ₹7,500/mo

	if _, err := pool.Exec(ctx, goalInsertQuery, userID, "Bike", 15000000, bikeDeadline, nil, 1); err != nil {
		logger.Error("failed to insert Bike goal", "error", err)
		os.Exit(1)
	}

	if _, err := pool.Exec(ctx, goalInsertQuery, userID, "MacBook Air", 9000000, nil, macbookMonthly, 2); err != nil {
		logger.Error("failed to insert MacBook goal", "error", err)
		os.Exit(1)
	}

	if _, err := pool.Exec(ctx, goalInsertQuery, userID, "Goa trip", 2500000, goaDeadline, nil, 3); err != nil {
		logger.Error("failed to insert Goa trip goal", "error", err)
		os.Exit(1)
	}
	logger.Info("seeded 3 savings goals with saved_paise = 0")

	// 12. Final confirmation banner
	fmt.Println()
	fmt.Println("==================================================================")
	fmt.Println("  Spendle Demo Data Seeded Successfully!")
	fmt.Println("==================================================================")
	fmt.Println("  Credentials:")
	fmt.Printf("    Email:    %s\n", demoEmail)
	fmt.Printf("    Password: %s\n", demoPassword)
	fmt.Println()
	fmt.Println("  Next steps for your demonstration:")
	fmt.Println("    1. Start Go Backend:  go run ./cmd/api")
	fmt.Println("       (With WORKER_INTERVAL=10s, watch it sweep M1 automatically!)")
	fmt.Println("    2. Start Frontend:    npm run dev (inside /frontend)")
	fmt.Println("    3. Log in at:         http://localhost:5173")
	fmt.Println("    4. Explore:")
	fmt.Println("       - Notice 'In profit' status banner and monthly totals")
	fmt.Println("       - Check 'Online Food' budget in M1 (visibly overspent in coral)")
	fmt.Println("       - Use '+ Expense' / '+ Income' floating sheet to log transactions")
	fmt.Println("       - Visit '/goals' to trigger manual month-close and watch savings grow")
	fmt.Println("==================================================================")
	fmt.Println()
}
