package auth

import (
	"context"
	"time"
)

// Repository defines the persistence interface contract for users, sessions, and OTP challenges.
// Go Concept - Interfaces:
// Defining an interface decouples our service logic from pgx/PostgreSQL, enabling
// comprehensive unit testing via an in-memory mock repository without a real database.
type Repository interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	VerifyUserEmail(ctx context.Context, email string) error
	UpdateUserPassword(ctx context.Context, email, passwordHash string) error
	UpsertGoogleUser(ctx context.Context, googleID, email, name string) (User, error)

	CreateOTP(ctx context.Context, email, codeHash, purpose string, maxAttempts int, expiresAt time.Time) (OTPChallenge, error)
	GetLatestActiveOTP(ctx context.Context, email, purpose string) (OTPChallenge, error)
	IncrementOTPAttempts(ctx context.Context, id string) error
	ConsumeOTP(ctx context.Context, id string) error
}
