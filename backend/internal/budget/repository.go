package budget

import (
	"context"
	"time"
)

// Repository defines the persistence interface contract for categories and budgets.
// Go Concept - Interface Driven Design:
// Services depend strictly on this interface. This guarantees that business logic can be
// unit-tested with lightweight in-memory fake repositories without requiring a live PostgreSQL instance.
type Repository interface {
	// ListCategories returns all categories belonging to the user sorted alphabetically by name.
	ListCategories(ctx context.Context, userID string) ([]Category, error)

	// CreateCategory inserts a new custom category for the user.
	CreateCategory(ctx context.Context, userID, name string) (Category, error)

	// DeleteCategory removes a category if it belongs to the user.
	DeleteCategory(ctx context.Context, userID, categoryID string) error

	// GetCategoryByID looks up a category ensuring it belongs to the authenticated user.
	GetCategoryByID(ctx context.Context, userID, categoryID string) (Category, error)

	// UpsertBudget atomically sets or updates a category limit for a specific 1st-of-month date.
	UpsertBudget(ctx context.Context, userID, categoryID string, month time.Time, limitPaise int64) (MonthlyBudget, error)

	// GetBudgetProgress returns single-query aggregated spent/left progress for all budgets in a month.
	GetBudgetProgress(ctx context.Context, userID string, monthStart, nextMonthStart time.Time) ([]BudgetProgress, error)
}
