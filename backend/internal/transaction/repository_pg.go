package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PGRepository implements Repository using PostgreSQL through pgxpool.
type PGRepository struct {
	pool *pgxpool.Pool
}

// NewPGRepository constructs a PostgreSQL transaction repository.
func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// CreateTransaction inserts a new income or expense transaction.
func (r *PGRepository) CreateTransaction(ctx context.Context, tx Transaction) (Transaction, error) {
	if r.pool == nil {
		return Transaction{}, errors.New("database connection pool is not initialized")
	}

	// Go Concept - User ID Scoping:
	// Every single SQL mutation or read is strictly scoped with `user_id = $1`.
	// Since user_id comes directly from the cryptographically verified JWT context,
	// this eliminates any possibility of horizontal privilege escalation (IDOR).
	query := `
		INSERT INTO transactions (user_id, type, amount_paise, category_id, payment_mode, note, occurred_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, type, amount_paise, category_id, payment_mode, COALESCE(note, ''), to_char(occurred_on, 'YYYY-MM-DD'), created_at;
	`

	var created Transaction
	err := r.pool.QueryRow(ctx, query,
		tx.UserID,
		tx.Type,
		tx.AmountPaise,
		tx.CategoryID,
		tx.PaymentMode,
		tx.Note,
		tx.OccurredOn,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Type,
		&created.AmountPaise,
		&created.CategoryID,
		&created.PaymentMode,
		&created.Note,
		&created.OccurredOn,
		&created.CreatedAt,
	)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to insert transaction: %w", err)
	}

	return created, nil
}

// ListTransactions retrieves transactions for a user within a calendar month ordered newest first.
// Go Concept - Nil vs Empty Slices:
// If a user has no transactions for the requested month, we return an initialized slice `make([]Transaction, 0)`.
// This guarantees that Go's JSON encoder serializes `[]` rather than `null`.
func (r *PGRepository) ListTransactions(ctx context.Context, userID string, monthStart, nextMonthStart time.Time, limit int) ([]Transaction, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT
			t.id,
			t.user_id,
			t.type,
			t.amount_paise,
			t.category_id,
			c.name,
			t.payment_mode,
			COALESCE(t.note, ''),
			to_char(t.occurred_on, 'YYYY-MM-DD'),
			t.created_at
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1
		  AND t.occurred_on >= $2
		  AND t.occurred_on < $3
		ORDER BY t.occurred_on DESC, t.created_at DESC
		LIMIT $4;
	`

	rows, err := r.pool.Query(ctx, query, userID, monthStart, nextMonthStart, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Type,
			&tx.AmountPaise,
			&tx.CategoryID,
			&tx.CategoryName,
			&tx.PaymentMode,
			&tx.Note,
			&tx.OccurredOn,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// DeleteTransaction removes a transaction scoped by user ID, returning ErrTransactionNotFound if missing.
func (r *PGRepository) DeleteTransaction(ctx context.Context, userID, transactionID string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2;
	`

	tag, err := r.pool.Exec(ctx, query, transactionID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

// GetMonthlySummary calculates total income, total expenses, and net profit/loss for the month.
// Go Concept - SQL Aggregate FILTER Clause:
// PostgreSQL supports `SUM(...) FILTER (WHERE condition)`.
// This allows conditional aggregation in a single query pass across the table,
// calculating both total income and total expenses simultaneously without multiple queries or subselects.
func (r *PGRepository) GetMonthlySummary(ctx context.Context, userID string, monthStart, nextMonthStart time.Time) (Summary, error) {
	if r.pool == nil {
		return Summary{}, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT
			COALESCE(SUM(amount_paise) FILTER (WHERE type = 'income'), 0) AS income_paise,
			COALESCE(SUM(amount_paise) FILTER (WHERE type = 'expense'), 0) AS expense_paise
		FROM transactions
		WHERE user_id = $1
		  AND occurred_on >= $2
		  AND occurred_on < $3;
	`

	var s Summary
	err := r.pool.QueryRow(ctx, query, userID, monthStart, nextMonthStart).Scan(
		&s.IncomePaise,
		&s.ExpensePaise,
	)
	if err != nil {
		return Summary{}, fmt.Errorf("failed to query monthly transaction summary: %w", err)
	}

	s.NetPaise = s.IncomePaise - s.ExpensePaise
	return s, nil
}

