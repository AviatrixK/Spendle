package goal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGRepository implements the Repository interface using PostgreSQL and pgxpool.
type PGRepository struct {
	pool *pgxpool.Pool
}

// NewPGRepository creates a new PostgreSQL goal repository instance.
func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// CreateGoal inserts a new savings goal for the user.
// Go Concept - User ID Scoping:
// Every query is explicitly scoped with user_id = $1 from request context.
func (r *PGRepository) CreateGoal(ctx context.Context, g Goal) (Goal, error) {
	if r.pool == nil {
		return Goal{}, errors.New("database connection pool is not initialized")
	}

	query := `
		INSERT INTO goals (user_id, title, target_paise, saved_paise, deadline, monthly_amount_paise, priority, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, title, target_paise, saved_paise, to_char(deadline, 'YYYY-MM-DD'), monthly_amount_paise, priority, status, created_at, updated_at;
	`

	var created Goal
	err := r.pool.QueryRow(ctx, query,
		g.UserID,
		g.Title,
		g.TargetPaise,
		g.SavedPaise,
		g.Deadline,
		g.MonthlyAmountPaise,
		g.Priority,
		g.Status,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Title,
		&created.TargetPaise,
		&created.SavedPaise,
		&created.Deadline,
		&created.MonthlyAmountPaise,
		&created.Priority,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return Goal{}, fmt.Errorf("failed to insert goal: %w", err)
	}

	return created, nil
}

// GetGoalByID retrieves a goal ensuring it belongs to the authenticated user.
func (r *PGRepository) GetGoalByID(ctx context.Context, userID, goalID string) (Goal, error) {
	if r.pool == nil {
		return Goal{}, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT id, user_id, title, target_paise, saved_paise, to_char(deadline, 'YYYY-MM-DD'), monthly_amount_paise, priority, status, created_at, updated_at
		FROM goals
		WHERE id = $1 AND user_id = $2;
	`

	var g Goal
	err := r.pool.QueryRow(ctx, query, goalID, userID).Scan(
		&g.ID,
		&g.UserID,
		&g.Title,
		&g.TargetPaise,
		&g.SavedPaise,
		&g.Deadline,
		&g.MonthlyAmountPaise,
		&g.Priority,
		&g.Status,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Goal{}, ErrGoalNotFound
		}
		return Goal{}, fmt.Errorf("failed to query goal: %w", err)
	}

	return g, nil
}

// ListGoals returns all goals belonging to the user ordered by priority ASC, created_at ASC.
func (r *PGRepository) ListGoals(ctx context.Context, userID string) ([]Goal, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT id, user_id, title, target_paise, saved_paise, to_char(deadline, 'YYYY-MM-DD'), monthly_amount_paise, priority, status, created_at, updated_at
		FROM goals
		WHERE user_id = $1
		ORDER BY priority ASC, created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list goals: %w", err)
	}
	defer rows.Close()

	goals := make([]Goal, 0)
	for rows.Next() {
		var g Goal
		if err := rows.Scan(
			&g.ID,
			&g.UserID,
			&g.Title,
			&g.TargetPaise,
			&g.SavedPaise,
			&g.Deadline,
			&g.MonthlyAmountPaise,
			&g.Priority,
			&g.Status,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan goal row: %w", err)
		}
		goals = append(goals, g)
	}

	return goals, rows.Err()
}

// UpdateGoal edits title, target_paise, deadline, monthly_amount_paise, priority, and status.
func (r *PGRepository) UpdateGoal(ctx context.Context, g Goal) (Goal, error) {
	if r.pool == nil {
		return Goal{}, errors.New("database connection pool is not initialized")
	}

	query := `
		UPDATE goals
		SET title = $1,
		    target_paise = $2,
		    deadline = $3,
		    monthly_amount_paise = $4,
		    priority = $5,
		    status = $6,
		    updated_at = now()
		WHERE id = $7 AND user_id = $8
		RETURNING id, user_id, title, target_paise, saved_paise, to_char(deadline, 'YYYY-MM-DD'), monthly_amount_paise, priority, status, created_at, updated_at;
	`

	var updated Goal
	err := r.pool.QueryRow(ctx, query,
		g.Title,
		g.TargetPaise,
		g.Deadline,
		g.MonthlyAmountPaise,
		g.Priority,
		g.Status,
		g.ID,
		g.UserID,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Title,
		&updated.TargetPaise,
		&updated.SavedPaise,
		&updated.Deadline,
		&updated.MonthlyAmountPaise,
		&updated.Priority,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Goal{}, ErrGoalNotFound
		}
		return Goal{}, fmt.Errorf("failed to update goal: %w", err)
	}

	return updated, nil
}

// DeleteGoal removes a goal belonging to the authenticated user.
func (r *PGRepository) DeleteGoal(ctx context.Context, userID, goalID string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		DELETE FROM goals
		WHERE id = $1 AND user_id = $2;
	`

	tag, err := r.pool.Exec(ctx, query, goalID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrGoalNotFound
	}

	return nil
}

// GetMaxPriority returns the highest priority number for this user, or 0 if no goals exist.
func (r *PGRepository) GetMaxPriority(ctx context.Context, userID string) (int, error) {
	if r.pool == nil {
		return 0, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT COALESCE(MAX(priority), 0)
		FROM goals
		WHERE user_id = $1;
	`

	var maxP int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&maxP)
	if err != nil {
		return 0, fmt.Errorf("failed to query max priority: %w", err)
	}

	return maxP, nil
}

// CloseMonth performs the month-end budget sweep inside ONE atomic database transaction.
// Go Concept - Database Transactions (ACID):
// Using Begin() and defer tx.Rollback() guarantees that either all rollover updates commit
// atomically or none do, avoiding any inconsistent financial state if an error occurs midway.
func (r *PGRepository) CloseMonth(
	ctx context.Context,
	userID string,
	month time.Time,
	allocateFn func(leftover int64, goals []Goal) []Allocation,
) (CloseResult, error) {
	if r.pool == nil {
		return CloseResult{}, errors.New("database connection pool is not initialized")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return CloseResult{}, fmt.Errorf("failed to begin month close transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Idempotency check with ON CONFLICT DO NOTHING
	// Go Concept - Idempotency:
	// A month closing operation must be safe to execute multiple times. If the user or background
	// worker calls CloseMonth again for the same month, 0 rows are inserted into month_closings.
	// We detect this via RowsAffected() == 0 and immediately return AlreadyClosed: true without
	// making any changes or double-counting leftover funds.
	insertClosingQuery := `
		INSERT INTO month_closings (user_id, month)
		VALUES ($1, $2)
		ON CONFLICT (user_id, month) DO NOTHING;
	`
	tag, err := tx.Exec(ctx, insertClosingQuery, userID, month)
	if err != nil {
		return CloseResult{}, fmt.Errorf("failed to insert month_closings: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return CloseResult{AlreadyClosed: true}, nil
	}

	// Step 2: Calculate leftover funds
	// leftover = (sum of this user's monthly_budgets.limit_paise for that month) -
	//            (sum of that user's EXPENSE transactions in that month, in budgeted categories only)
	budgetSumQuery := `
		SELECT COALESCE(SUM(limit_paise), 0)
		FROM monthly_budgets
		WHERE user_id = $1 AND month = $2;
	`
	var totalLimit int64
	if err := tx.QueryRow(ctx, budgetSumQuery, userID, month).Scan(&totalLimit); err != nil {
		return CloseResult{}, fmt.Errorf("failed to calculate monthly budget limits: %w", err)
	}

	nextMonth := month.AddDate(0, 1, 0)
	expenseSumQuery := `
		SELECT COALESCE(SUM(t.amount_paise), 0)
		FROM transactions t
		WHERE t.user_id = $1
		  AND t.type = 'expense'
		  AND t.occurred_on >= $2
		  AND t.occurred_on < $3
		  AND t.category_id IN (
		      SELECT category_id FROM monthly_budgets WHERE user_id = $1 AND month = $2
		  );
	`
	var totalExpense int64
	if err := tx.QueryRow(ctx, expenseSumQuery, userID, month, nextMonth).Scan(&totalExpense); err != nil {
		return CloseResult{}, fmt.Errorf("failed to calculate budgeted expenses: %w", err)
	}

	leftover := totalLimit - totalExpense
	if leftover <= 0 {
		// Commit transaction to finalize month_closings record, ensuring month isn't scanned again
		if err := tx.Commit(ctx); err != nil {
			return CloseResult{}, fmt.Errorf("failed to commit empty month closing: %w", err)
		}
		return CloseResult{AlreadyClosed: false, MovedPaise: 0, GoalsFunded: 0}, nil
	}

	// Step 3: Load active goals ordered by priority
	activeGoalsQuery := `
		SELECT id, user_id, title, target_paise, saved_paise, to_char(deadline, 'YYYY-MM-DD'), monthly_amount_paise, priority, status, created_at, updated_at
		FROM goals
		WHERE user_id = $1 AND status = 'active'
		ORDER BY priority ASC, created_at ASC;
	`
	rows, err := tx.Query(ctx, activeGoalsQuery, userID)
	if err != nil {
		return CloseResult{}, fmt.Errorf("failed to query active goals for sweep: %w", err)
	}

	var activeGoals []Goal
	for rows.Next() {
		var g Goal
		if err := rows.Scan(
			&g.ID,
			&g.UserID,
			&g.Title,
			&g.TargetPaise,
			&g.SavedPaise,
			&g.Deadline,
			&g.MonthlyAmountPaise,
			&g.Priority,
			&g.Status,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			rows.Close()
			return CloseResult{}, fmt.Errorf("failed to scan active goal row: %w", err)
		}
		activeGoals = append(activeGoals, g)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return CloseResult{}, fmt.Errorf("active goals rows error: %w", err)
	}

	// Go Concept - Pure Function Invocation via Function Parameter:
	// allocateFn is passed into CloseMonth, decoupling database operations from the business rule.
	allocations := allocateFn(leftover, activeGoals)

	var movedPaise int64
	var goalsFunded int

	for _, alloc := range allocations {
		if alloc.AmountPaise <= 0 {
			continue
		}

		// Update goal balance and status
		updateGoalQuery := `
			UPDATE goals
			SET saved_paise = saved_paise + $1,
			    status = CASE WHEN saved_paise + $1 >= target_paise THEN 'completed' ELSE status END,
			    updated_at = now()
			WHERE id = $2 AND user_id = $3;
		`
		if _, err := tx.Exec(ctx, updateGoalQuery, alloc.AmountPaise, alloc.GoalID, userID); err != nil {
			return CloseResult{}, fmt.Errorf("failed to update goal %s during rollover: %w", alloc.GoalID, err)
		}

		// Insert goal contribution audit row
		insertContribQuery := `
			INSERT INTO goal_contributions (goal_id, user_id, amount_paise, source, month)
			VALUES ($1, $2, $3, 'rollover', $4);
		`
		if _, err := tx.Exec(ctx, insertContribQuery, alloc.GoalID, userID, alloc.AmountPaise, month); err != nil {
			return CloseResult{}, fmt.Errorf("failed to insert contribution for goal %s: %w", alloc.GoalID, err)
		}

		movedPaise += alloc.AmountPaise
		goalsFunded++
	}

	// Step 4: Commit transaction and return result
	if err := tx.Commit(ctx); err != nil {
		return CloseResult{}, fmt.Errorf("failed to commit month closing transaction: %w", err)
	}

	return CloseResult{
		AlreadyClosed: false,
		MovedPaise:    movedPaise,
		GoalsFunded:   goalsFunded,
	}, nil
}

// GetUnclosedPreviousMonths retrieves (user_id, previous_month) jobs needing rollover.
func (r *PGRepository) GetUnclosedPreviousMonths(ctx context.Context, previousMonth time.Time) ([]RolloverJob, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT DISTINCT b.user_id, b.month
		FROM monthly_budgets b
		LEFT JOIN month_closings mc ON mc.user_id = b.user_id AND mc.month = b.month
		WHERE b.month = $1 AND mc.user_id IS NULL;
	`

	rows, err := r.pool.Query(ctx, query, previousMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to query unclosed previous months: %w", err)
	}
	defer rows.Close()

	jobs := make([]RolloverJob, 0)
	for rows.Next() {
		var j RolloverJob
		if err := rows.Scan(&j.UserID, &j.Month); err != nil {
			return nil, fmt.Errorf("failed to scan rollover job: %w", err)
		}
		jobs = append(jobs, j)
	}

	return jobs, rows.Err()
}

