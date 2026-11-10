package transaction

import (
	"context"
	"time"
)

// Repository defines the database operations contract for transactions.
// Go Concept - Interface Driven Architecture:
// High-level service modules depend on this interface abstraction rather than
// concrete PostgreSQL drivers, ensuring straightforward in-memory unit testing.
type Repository interface {
	// CreateTransaction inserts a new income or expense transaction for the user.
	CreateTransaction(ctx context.Context, tx Transaction) (Transaction, error)

	// ListTransactions retrieves transactions for a user in a date range ordered newest first.
	ListTransactions(ctx context.Context, userID string, monthStart, nextMonthStart time.Time, limit int) ([]Transaction, error)

	// DeleteTransaction deletes a transaction if it belongs to the user.
	DeleteTransaction(ctx context.Context, userID, transactionID string) error

	// GetMonthlySummary calculates total income, expense, and net paise using an aggregated SQL query.
	GetMonthlySummary(ctx context.Context, userID string, monthStart, nextMonthStart time.Time) (Summary, error)
}

