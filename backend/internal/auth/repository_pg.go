package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// CreateUser inserts an unverified user and initializes their 8 default expense categories in a single atomic transaction.
func (r *PGRepository) CreateUser(ctx context.Context, name, email, passwordHash string) (User, error) {
	if r.pool == nil {
		return User{}, errors.New("database connection pool is not initialized")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var user User
	userQuery := `
		INSERT INTO users (name, email, password_hash, is_verified, token_version)
		VALUES ($1, $2, $3, false, 1)
		RETURNING id, name, email, password_hash, is_verified, verified_at, token_version, google_id, created_at;
	`
	err = tx.QueryRow(ctx, userQuery, name, email, passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.VerifiedAt,
		&user.TokenVersion,
		&user.GoogleID,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("failed to insert user: %w", err)
	}

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
		SELECT id, name, email, password_hash, is_verified, verified_at, token_version, google_id, created_at
		FROM users
		WHERE email = $1;
	`

	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.VerifiedAt,
		&user.TokenVersion,
		&user.GoogleID,
		&user.CreatedAt,
	)
	if err != nil {
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
		SELECT id, name, email, password_hash, is_verified, verified_at, token_version, google_id, created_at
		FROM users
		WHERE id = $1;
	`

	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.VerifiedAt,
		&user.TokenVersion,
		&user.GoogleID,
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

// VerifyUserEmail marks the user's email as verified and stamps verified_at.
func (r *PGRepository) VerifyUserEmail(ctx context.Context, email string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		UPDATE users
		SET is_verified = true, verified_at = now()
		WHERE email = $1;
	`
	tag, err := r.pool.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("failed to mark email verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserPassword updates password hash and increments token_version to invalidate prior JWT sessions.
func (r *PGRepository) UpdateUserPassword(ctx context.Context, email, passwordHash string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `
		UPDATE users
		SET password_hash = $1, token_version = token_version + 1
		WHERE email = $2;
	`
	tag, err := r.pool.Exec(ctx, query, passwordHash, email)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpsertGoogleUser links an existing account to Google or provisions a new verified user with default categories.
func (r *PGRepository) UpsertGoogleUser(ctx context.Context, googleID, email, name string) (User, error) {
	if r.pool == nil {
		return User{}, errors.New("database connection pool is not initialized")
	}

	// 1. Check if user already exists by email
	if _, err := r.GetUserByEmail(ctx, email); err == nil {
		// Link Google ID and mark verified if not already
		updateQuery := `
			UPDATE users
			SET google_id = $1, is_verified = true, verified_at = COALESCE(verified_at, now())
			WHERE email = $2
			RETURNING id, name, email, password_hash, is_verified, verified_at, token_version, google_id, created_at;
		`
		var updated User
		err = r.pool.QueryRow(ctx, updateQuery, googleID, email).Scan(
			&updated.ID,
			&updated.Name,
			&updated.Email,
			&updated.PasswordHash,
			&updated.IsVerified,
			&updated.VerifiedAt,
			&updated.TokenVersion,
			&updated.GoogleID,
			&updated.CreatedAt,
		)
		if err != nil {
			return User{}, fmt.Errorf("failed to link Google identity: %w", err)
		}
		return updated, nil
	}

	// 2. User does not exist, provision new user in transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var user User
	insertQuery := `
		INSERT INTO users (name, email, password_hash, is_verified, verified_at, token_version, google_id)
		VALUES ($1, $2, NULL, true, now(), 1, $3)
		RETURNING id, name, email, password_hash, is_verified, verified_at, token_version, google_id, created_at;
	`
	err = tx.QueryRow(ctx, insertQuery, name, email, googleID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.VerifiedAt,
		&user.TokenVersion,
		&user.GoogleID,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("failed to create Google user: %w", err)
	}

	defaultCategories := []string{
		"Food", "Groceries", "Transport", "Bills",
		"Shopping", "Entertainment", "Health", "Other",
	}
	categoryQuery := `INSERT INTO categories (user_id, name) VALUES ($1, $2);`
	for _, cat := range defaultCategories {
		if _, err := tx.Exec(ctx, categoryQuery, user.ID, cat); err != nil {
			return User{}, fmt.Errorf("failed to insert default category %q: %w", cat, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("failed to commit Google signup transaction: %w", err)
	}

	return user, nil
}

// CreateOTP invalidates prior unconsumed OTPs for the same email/purpose, then stores the new challenge.
func (r *PGRepository) CreateOTP(ctx context.Context, email, codeHash, purpose string, maxAttempts int, expiresAt time.Time) (OTPChallenge, error) {
	if r.pool == nil {
		return OTPChallenge{}, errors.New("database connection pool is not initialized")
	}

	// Invalidate any existing active OTP for this email and purpose
	invalidateQuery := `
		UPDATE otps
		SET consumed_at = now()
		WHERE email = $1 AND purpose = $2 AND consumed_at IS NULL;
	`
	if _, err := r.pool.Exec(ctx, invalidateQuery, email, purpose); err != nil {
		return OTPChallenge{}, fmt.Errorf("failed to invalidate prior OTPs: %w", err)
	}

	insertQuery := `
		INSERT INTO otps (email, code_hash, purpose, attempts, max_attempts, expires_at)
		VALUES ($1, $2, $3, 0, $4, $5)
		RETURNING id, email, code_hash, purpose, attempts, max_attempts, expires_at, consumed_at, created_at;
	`

	var challenge OTPChallenge
	err := r.pool.QueryRow(ctx, insertQuery, email, codeHash, purpose, maxAttempts, expiresAt).Scan(
		&challenge.ID,
		&challenge.Email,
		&challenge.CodeHash,
		&challenge.Purpose,
		&challenge.Attempts,
		&challenge.MaxAttempts,
		&challenge.ExpiresAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	)
	if err != nil {
		return OTPChallenge{}, fmt.Errorf("failed to create OTP challenge: %w", err)
	}

	return challenge, nil
}

// GetLatestActiveOTP retrieves the most recently created, unconsumed OTP for the email and purpose.
func (r *PGRepository) GetLatestActiveOTP(ctx context.Context, email, purpose string) (OTPChallenge, error) {
	if r.pool == nil {
		return OTPChallenge{}, errors.New("database connection pool is not initialized")
	}

	query := `
		SELECT id, email, code_hash, purpose, attempts, max_attempts, expires_at, consumed_at, created_at
		FROM otps
		WHERE email = $1 AND purpose = $2 AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1;
	`

	var challenge OTPChallenge
	err := r.pool.QueryRow(ctx, query, email, purpose).Scan(
		&challenge.ID,
		&challenge.Email,
		&challenge.CodeHash,
		&challenge.Purpose,
		&challenge.Attempts,
		&challenge.MaxAttempts,
		&challenge.ExpiresAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OTPChallenge{}, ErrNotFound
		}
		return OTPChallenge{}, fmt.Errorf("failed to query active OTP: %w", err)
	}

	return challenge, nil
}

// IncrementOTPAttempts increments the failed verification counter for the challenge.
func (r *PGRepository) IncrementOTPAttempts(ctx context.Context, id string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `UPDATE otps SET attempts = attempts + 1 WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// ConsumeOTP marks the OTP challenge as consumed/used so it can never be reused.
func (r *PGRepository) ConsumeOTP(ctx context.Context, id string) error {
	if r.pool == nil {
		return errors.New("database connection pool is not initialized")
	}

	query := `UPDATE otps SET consumed_at = now() WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
