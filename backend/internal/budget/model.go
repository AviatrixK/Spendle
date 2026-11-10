package budget

import "time"

// Category represents a user expense classification (e.g., Food, Transport, Bills).
// Go Concept - Struct Tags:
// The backtick annotations like `json:"id"` are struct tags. They instruct Go's encoding/json
// package how to map struct field names to JSON keys during serialization and deserialization.
type Category struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// CategoryResponse represents the public JSON shape for category listings: [{id, name}].
type CategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MonthlyBudget represents a category's configured spending limit for a specific month.
// Strict Monetary Precision: All budget limits are stored strictly in integer paise (int64).
type MonthlyBudget struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CategoryID string    `json:"category_id"`
	Month      string    `json:"month"` // Formatted as YYYY-MM
	LimitPaise int64     `json:"limit_paise"`
	CreatedAt  time.Time `json:"created_at"`
}

// BudgetProgress represents a category's spent vs left breakdown.
type BudgetProgress struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	LimitPaise   int64  `json:"limit_paise"`
	SpentPaise   int64  `json:"spent_paise"`
	LeftPaise    int64  `json:"left_paise"`
}

// MonthBudgetSummary contains the top-level monthly totals and category progress rows.
type MonthBudgetSummary struct {
	Month           string           `json:"month"`
	TotalLimitPaise int64            `json:"total_limit_paise"`
	TotalSpentPaise int64            `json:"total_spent_paise"`
	TotalLeftPaise  int64            `json:"total_left_paise"`
	Budgets         []BudgetProgress `json:"budgets"`
}
