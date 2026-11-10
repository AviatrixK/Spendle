package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Precomputed valid bcrypt hash (cost 12) used to equalize login response timing.
// When an unrecognized email attempts to log in, we still execute bcrypt against this dummy hash
// to ensure the request takes the exact same computation time (~100-250ms), preventing timing attacks.
const dummyBcryptHash = "$2a$12$4mUqgQeJ9D3y1vA8F6d8.uWpW2QoQeJ9D3y1vA8F6d8.uWpW2QoQe"

// Service handles authentication business logic: registration, credential validation,
// password hashing, and token issuance.
type Service struct {
	repo     Repository
	tokenMgr *TokenManager
}

// NewService constructs an auth Service with repository and token manager dependencies.
func NewService(repo Repository, tokenMgr *TokenManager) *Service {
	return &Service{
		repo:     repo,
		tokenMgr: tokenMgr,
	}
}

// Signup validates registration inputs, hashes the password with bcrypt (cost 12),
// persists the user and their default categories, and issues an authentication JWT.
func (s *Service) Signup(ctx context.Context, name, email, password string) (User, string, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 60 {
		return User{}, "", errors.New("name must be between 2 and 60 characters")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmail(email) {
		return User{}, "", errors.New("invalid email address format")
	}

	// Bcrypt has an architectural maximum input length of 72 bytes.
	// Input longer than 72 bytes is silently truncated by bcrypt, making 73+ byte passwords misleading.
	if len(password) < 8 || len(password) > 72 {
		return User{}, "", errors.New("password must be between 8 and 72 characters")
	}

	// Hash password using bcrypt at cost 12
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, name, email, string(hashBytes))
	if err != nil {
		return User{}, "", err // May be ErrEmailTaken or DB error
	}

	token, err := s.tokenMgr.GenerateToken(user.ID)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// Login verifies credentials using constant-time comparison techniques and returns an authentication JWT.
func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return User{}, "", ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Go Concept - Timing Attack Mitigation:
			// Bcrypt is deliberately computationally intensive. If we return an immediate error
			// when an email does not exist, the response takes <5ms. When the email exists, checking
			// the password takes ~150ms. Attackers can measure response times to enumerate valid user accounts.
			// By executing a dummy bcrypt comparison, non-existent accounts take the exact same time as real ones.
			_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(password))
			return User{}, "", ErrInvalidCredentials
		}
		return User{}, "", err
	}

	// Compare provided password against stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := s.tokenMgr.GenerateToken(user.ID)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// Me retrieves the authenticated user's profile by their unique ID.
func (s *Service) Me(ctx context.Context, userID string) (User, error) {
	if userID == "" {
		return User{}, ErrInvalidToken
	}
	return s.repo.GetUserByID(ctx, userID)
}

// isValidEmail performs RFC 5322 validation and ensures basic domain shape.
func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

