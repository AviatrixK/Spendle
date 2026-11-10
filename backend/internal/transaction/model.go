package transaction

import "time"

// Transaction represents an individual financial transaction (income or expense).
// Strict Financial Precision: All transaction amounts are stored strictly in integer paise (int64).
type Transaction struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Type         string    `json:"type"` // "income" or "expense"
	AmountPaise  int64     `json:"amount_paise"`
	CategoryID   *string   `json:"category_id"`
	CategoryName *string   `json:"category_name"`
	PaymentMode  string    `json:"payment_mode"` // "upi", "cash", "card", "netbanking", "other"
	Note         string    `json:"note"`
	OccurredOn   string    `json:"occurred_on"` // Formatted as YYYY-MM-DD
	CreatedAt    time.Time `json:"created_at"`
}

// Summary represents the aggregated financial metrics for a calendar month.
type Summary struct {
	IncomePaise  int64 `json:"income_paise"`
	ExpensePaise int64 `json:"expense_paise"`
	NetPaise     int64 `json:"net_paise"` // Income - Expense (can be negative, indicating net loss)
}

