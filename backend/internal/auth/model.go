package auth

import "time"

// User represents the domain model for an authenticated user account.
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Must NEVER appear in any JSON response
	CreatedAt    time.Time `json:"created_at"`
}

// UserResponse is the safe, sanitized representation of a user returned in HTTP responses.
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a User entity into a safe UserResponse.
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

