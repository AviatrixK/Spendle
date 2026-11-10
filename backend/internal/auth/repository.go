package auth

import "context"

// Go Concept - Interfaces:
// In Go, an interface defines a contract of behavior (method signatures).
// Types implement interfaces implicitly—there is no explicit `implements` keyword.
// By having the auth Service depend on this `Repository` interface rather than a concrete
// PostgreSQL struct, we can swap in an in-memory mock repository for lightning-fast unit tests
// without needing a real database connection.
type Repository interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
}

