package budget

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGRepository implements Repository with PostgreSQL using pgxpool.
type PGRepository struct {
	pool *pgxpool.Pool
}

// NewPGRepository constructs a PostgreSQL budget repository instance.
func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// ListCategories retrieves all categories for a user sorted alphabetically.
// Go Concept - Nil vs Empty Slices:
// If a user has no categories, a nil slice (`var list []Category`) would serialize into JSON
// as `null`. Initializing with `make([]Category, 0)` ensures it encodes as an empty array `[]`.
func (r *PGRepository) ListCategories(ctx context.Context, userID string) ([]Category, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	// Go Concept - User ID Scoping:
	// Every single SQL query MUST include `WHERE user_id = $1`.
	// Because the user ID is securely extracted from the validated JWT in context (not client input),
	// this guarantees strict multi-tenant isolation: a user can never inspect or alter another user's records.
	query := `
		SELECT id, name
		FROM categories
		WHERE user_id = $1
		ORDER BY name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// CreateCategory persists a custom category, returning ErrCategoryExists on duplicate name.
func (r *PGRepository) CreateCategory(ctx context.Context, userID, name string) (Category, error) {
	if r.pool == nil {
		return Category{}, errors.New("database connection pool is not initialized")
	}

	query := `
		INSERT INTO categories (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at;
	`

	var c Category
	err := r.pool.QueryRow(ctx, query, userID, name).Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Category{}, ErrCategoryExists
		}
		return Category{}, fmt.Errorf("failed to insert category: %w", err)
	}

	return c, nil
}

// DeleteCategory deletes a category scoped by user ID, returning ErrCategoryNotFound if 0 rows matched.
func (r *PGRepository) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		DELETE FROM categories
		WHERE id = $1 AND user_id = $2;
	`

	tag, err := r.pool.Exec(ctx, query, categoryID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// GetCategoryByID looks up a category ensuring ownership.
func (r *PGRepository) GetCategoryByID(ctx context.Context, userID, categoryID string) (Category, error) {
	if r.pool == nil {
		return Category{}, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT id, user_id, name, created_at
		FROM categories
		WHERE id = $1 AND user_id = $2;
	`

	var c Category
	err := r.pool.QueryRow(ctx, query, categoryID, userID).Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("failed to query category: %w", err)
	}

	return c, nil
}

// UpsertBudget inserts or updates a category budget limit for a specific month.
func (r *PGRepository) UpsertBudget(ctx context.Context, userID, categoryID string, month time.Time, limitPaise int64) (MonthlyBudget, error) {
	if r.pool == nil {
		return MonthlyBudget{}, errors.New("database connection pool is not initialized")
	}

	query := `
		INSERT INTO monthly_budgets (user_id, category_id, month, limit_paise)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, category_id, month)
		DO UPDATE SET limit_paise = EXCLUDED.limit_paise
		RETURNING id, user_id, category_id, to_char(month, 'YYYY-MM'), limit_paise, created_at;
	`

	var b MonthlyBudget
	err := r.pool.QueryRow(ctx, query, userID, categoryID, month, limitPaise).Scan(
		&b.ID,
		&b.UserID,
		&b.CategoryID,
		&b.Month,
		&b.LimitPaise,
		&b.CreatedAt,
	)
	if err != nil {
		return MonthlyBudget{}, fmt.Errorf("failed to upsert budget: %w", err)
	}

	return b, nil
}

// GetBudgetProgress executes a single SQL query aggregating spent and left amounts.
// Go Concept - SQL Aggregation & COALESCE:
// Instead of fetching budgets and then running N queries in a loop to sum transactions,
// we join `monthly_budgets` with `categories` and `LEFT JOIN transactions`.
// `COALESCE(SUM(t.amount_paise), 0)` safely turns nulls (categories with 0 expenses) into 0 paise.
// This single roundtrip reduces database load and guarantees transactional consistency.
func (r *PGRepository) GetBudgetProgress(ctx context.Context, userID string, monthStart, nextMonthStart time.Time) ([]BudgetProgress, error) {
	if r.pool == nil {
		return nil, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT
			b.category_id,
			c.name AS category_name,
			b.limit_paise,
			COALESCE(SUM(t.amount_paise), 0) AS spent_paise,
			b.limit_paise - COALESCE(SUM(t.amount_paise), 0) AS left_paise
		FROM monthly_budgets b
		JOIN categories c ON b.category_id = c.id
		LEFT JOIN transactions t ON t.category_id = b.category_id
			AND t.user_id = b.user_id
			AND t.type = 'expense'
			AND t.occurred_on >= $2
			AND t.occurred_on < $3
		WHERE b.user_id = $1 AND b.month = $2
		GROUP BY b.category_id, c.name, b.limit_paise
		ORDER BY c.name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID, monthStart, nextMonthStart)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate budget progress: %w", err)
	}
	defer rows.Close()

	results := make([]BudgetProgress, 0)
	for rows.Next() {
		var bp BudgetProgress
		if err := rows.Scan(
			&bp.CategoryID,
			&bp.CategoryName,
			&bp.LimitPaise,
			&bp.SpentPaise,
			&bp.LeftPaise,
		); err != nil {
			return nil, fmt.Errorf("failed to scan budget progress row: %w", err)
		}
		results = append(results, bp)
	}

	return results, rows.Err()
}
