package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGRepository implements the Repository interface using PostgreSQL and pgxpool.
type PGRepository struct {
	pool *pgxpool.Pool
}

// NewPGRepository constructs a PostgreSQL-backed authentication repository.
func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

// CreateUser inserts a new user and creates their 8 default expense categories in a single atomic transaction.
func (r *PGRepository) CreateUser(ctx context.Context, name, email, passwordHash string) (User, error) {
	if r.pool == nil {
		return User{}, errors.New("database connection pool is not initialized")
	}

	// Go Concept - Database Transactions:
	// A transaction groups multiple SQL operations into an atomic unit (all succeed or all rollback).
	// When a user signs up, we must insert both the user row AND their default categories.
	// If inserting any category fails, `defer tx.Rollback` cancels the entire transaction,
	// preventing broken orphan accounts from existing in the database.
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) // Safe no-op if tx.Commit() succeeds
	}()

	var user User
	userQuery := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, created_at;
	`
	err = tx.QueryRow(ctx, userQuery, name, email, passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		// Go Concept - errors.As:
		// errors.As unwraps an error to see if it matches a specific type.
		// PostgreSQL error code 23505 indicates a unique constraint violation (e.g. email already exists).
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("failed to insert user: %w", err)
	}

	// Requirement: Create default categories for the new user
	defaultCategories := []string{
		"Food", "Groceries", "Transport", "Bills",
		"Shopping", "Entertainment", "Health", "Other",
	}

	categoryQuery := `
		INSERT INTO categories (user_id, name)
		VALUES ($1, $2);
	`
	for _, cat := range defaultCategories {
		if _, err := tx.Exec(ctx, categoryQuery, user.ID, cat); err != nil {
			return User{}, fmt.Errorf("failed to insert default category %q: %w", cat, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("failed to commit signup transaction: %w", err)
	}

	return user, nil
}

// GetUserByEmail queries a user by lowercase email address.
func (r *PGRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	if r.pool == nil {
		return User{}, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1;
	`

	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		// Go Concept - errors.Is:
		// errors.Is checks if an error matches a sentinel target anywhere in its chain.
		// We translate the driver-level pgx.ErrNoRows into our domain sentinel ErrNotFound.
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("failed to query user by email: %w", err)
	}

	return user, nil
}

// GetUserByID queries a user by their UUID primary key.
func (r *PGRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	if r.pool == nil {
		return User{}, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE id = $1;
	`

	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("failed to query user by id: %w", err)
	}

	return user, nil
}

