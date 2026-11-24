package auth

import "time"

// User represents the domain model for an authenticated user account.
type User struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash *string    `json:"-"` // Pointer to allow NULL for pure Google OAuth users
	IsVerified   bool       `json:"is_verified"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	TokenVersion int        `json:"-"` // Incremented upon password reset to revoke prior sessions
	GoogleID     *string    `json:"-"` // Optional Google OAuth identity mapping
	CreatedAt    time.Time  `json:"created_at"`
}

// UserResponse is the safe, sanitized representation of a user returned in HTTP responses.
type UserResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	IsVerified bool       `json:"is_verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToResponse converts a User entity into a safe UserResponse.
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		IsVerified: u.IsVerified,
		VerifiedAt: u.VerifiedAt,
		CreatedAt:  u.CreatedAt,
	}
}

// OTPChallenge represents a cryptographically hashed one-time passcode verification challenge.
type OTPChallenge struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	CodeHash    string     `json:"-"`
	Purpose     string     `json:"purpose"` // "email_verification" or "password_reset"
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	ExpiresAt   time.Time  `json:"expires_at"`
	ConsumedAt  *time.Time `json:"consumed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
