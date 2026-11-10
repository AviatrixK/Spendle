package auth

import "errors"

// Go Concept - Sentinel Errors:
// Sentinel errors are predefined, package-level error variables (conventionally starting with `Err`).
// Instead of inspecting error message strings, callers compare them directly using `errors.Is(err, auth.ErrNotFound)`.
// This decouples business logic from internal implementation details like SQL driver codes.

var (
	// ErrEmailTaken is returned when registration is attempted with an already existing email address.
	ErrEmailTaken = errors.New("email is already registered")

	// ErrNotFound is returned when a requested user does not exist in the database.
	ErrNotFound = errors.New("user not found")

	// ErrInvalidCredentials is intentionally returned for BOTH non-existent emails and incorrect passwords.
	// This prevents attackers from enumerating valid account emails (Account Enumeration Attack).
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrInvalidToken is returned when an authentication JWT is missing, malformed, expired, or invalid.
	ErrInvalidToken = errors.New("invalid or expired authentication token")
)

