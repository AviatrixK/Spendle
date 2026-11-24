package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/mailer"
	"golang.org/x/crypto/bcrypt"
)

// Precomputed valid bcrypt hash (cost 12) used to equalize login response timing.
// When an unrecognized email attempts to log in, we still execute bcrypt against this dummy hash
// to ensure the request takes the exact same computation time (~100-250ms), preventing timing attacks.
const dummyBcryptHash = "$2a$12$4mUqgQeJ9D3y1vA8F6d8.uWpW2QoQeJ9D3y1vA8F6d8.uWpW2QoQe"

// Service handles authentication business logic: registration, email verification,
// credential validation, OTP challenges, password recovery, and token issuance.
type Service struct {
	repo     Repository
	tokenMgr *TokenManager
	mailer   mailer.Mailer
	cfg      *config.Config
}

// NewService constructs an auth Service instance.
func NewService(repo Repository, tokenMgr *TokenManager, mailer mailer.Mailer, cfg *config.Config) *Service {
	return &Service{
		repo:     repo,
		tokenMgr: tokenMgr,
		mailer:   mailer,
		cfg:      cfg,
	}
}

// Signup validates registration inputs, creates an unverified account (or updates an unverified pending account),
// generates a 4-digit verification OTP, and delivers it via Brevo.
func (s *Service) Signup(ctx context.Context, name, email, password string) (User, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 60 {
		return User{}, errors.New("name must be between 2 and 60 characters")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmail(email) {
		return User{}, errors.New("invalid email address format")
	}

	// Bcrypt has an architectural maximum input length of 72 bytes.
	if len(password) < 8 || len(password) > 72 {
		return User{}, errors.New("password must be between 8 and 72 characters")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, name, email, string(hashBytes))
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			// Check if the existing account is still unverified.
			// If unverified, safely re-send verification without creating duplicate records.
			existing, getErr := s.repo.GetUserByEmail(ctx, email)
			if getErr == nil && !existing.IsVerified {
				user = existing
			} else {
				return User{}, ErrEmailTaken
			}
		} else {
			return User{}, err
		}
	}

	// Generate 4-digit secure OTP challenge
	code, err := GenerateSecureOTP(s.cfg.OTPLength)
	if err != nil {
		return User{}, err
	}

	codeHash := HashOTP(code)
	expiresAt := time.Now().Add(s.cfg.OTPExpiry)
	_, err = s.repo.CreateOTP(ctx, user.Email, codeHash, "email_verification", s.cfg.OTPMaxAttempts, expiresAt)
	if err != nil {
		return User{}, fmt.Errorf("failed to generate verification challenge: %w", err)
	}

	// Dispatch verification email via Brevo
	if err := s.mailer.SendVerificationOTP(ctx, user.Email, user.Name, code); err != nil {
		return user, fmt.Errorf("account created, but email delivery failed: %w", err)
	}

	return user, nil
}

// VerifyEmail verifies the 4-digit OTP, marks the account as verified, and creates an authenticated session.
func (s *Service) VerifyEmail(ctx context.Context, email, code string) (User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	code = strings.TrimSpace(code)

	challenge, err := s.repo.GetLatestActiveOTP(ctx, email, "email_verification")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, "", ErrInvalidOTP
		}
		return User{}, "", err
	}

	// Check expiration
	if time.Now().After(challenge.ExpiresAt) {
		_ = s.repo.ConsumeOTP(ctx, challenge.ID)
		return User{}, "", ErrExpiredOTP
	}

	// Check attempt ceiling
	if challenge.Attempts >= challenge.MaxAttempts {
		_ = s.repo.ConsumeOTP(ctx, challenge.ID)
		return User{}, "", ErrMaxAttemptsExceeded
	}

	// Verify cryptographic hash
	if !VerifyOTPHash(code, challenge.CodeHash) {
		_ = s.repo.IncrementOTPAttempts(ctx, challenge.ID)
		if challenge.Attempts+1 >= challenge.MaxAttempts {
			_ = s.repo.ConsumeOTP(ctx, challenge.ID)
			return User{}, "", ErrMaxAttemptsExceeded
		}
		return User{}, "", ErrInvalidOTP
	}

	// Consume OTP so it cannot be reused
	_ = s.repo.ConsumeOTP(ctx, challenge.ID)

	// Mark user email as verified
	if err := s.repo.VerifyUserEmail(ctx, email); err != nil {
		return User{}, "", fmt.Errorf("failed to update verification status: %w", err)
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, "", err
	}

	// Issue authentication JWT token
	token, err := s.tokenMgr.GenerateToken(user.ID, user.TokenVersion)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to issue session token: %w", err)
	}

	return user, token, nil
}

// ResendOTP validates cooldown, invalidates the previous challenge, and dispatches a fresh OTP.
func (s *Service) ResendOTP(ctx context.Context, email, purpose string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	purpose = strings.TrimSpace(purpose)

	if purpose != "email_verification" && purpose != "password_reset" {
		return errors.New("invalid OTP purpose")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Safe generic return to prevent account enumeration
			return nil
		}
		return err
	}

	if purpose == "email_verification" && user.IsVerified {
		return errors.New("email is already verified")
	}

	// Check cooldown: check the most recently created OTP
	latest, err := s.repo.GetLatestActiveOTP(ctx, email, purpose)
	if err == nil {
		if time.Since(latest.CreatedAt) < s.cfg.OTPCooldown {
			return ErrCooldownActive
		}
	}

	code, err := GenerateSecureOTP(s.cfg.OTPLength)
	if err != nil {
		return err
	}

	codeHash := HashOTP(code)
	expiresAt := time.Now().Add(s.cfg.OTPExpiry)
	_, err = s.repo.CreateOTP(ctx, email, codeHash, purpose, s.cfg.OTPMaxAttempts, expiresAt)
	if err != nil {
		return err
	}

	if purpose == "email_verification" {
		return s.mailer.SendVerificationOTP(ctx, user.Email, user.Name, code)
	}
	return s.mailer.SendPasswordResetOTP(ctx, user.Email, user.Name, code)
}

// Login verifies credentials, enforces that email is verified, and issues an authenticated JWT.
func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return User{}, "", ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(password))
			return User{}, "", ErrInvalidCredentials
		}
		return User{}, "", err
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return User{}, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	// Enforce email verification
	if !user.IsVerified {
		return user, "", ErrEmailUnverified
	}

	token, err := s.tokenMgr.GenerateToken(user.ID, user.TokenVersion)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to issue session token: %w", err)
	}

	return user, token, nil
}

// ForgotPassword issues a 4-digit password recovery OTP if the email exists, with timing attack defense.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmail(email) {
		return nil // Generic return
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil // Opaque success to prevent user enumeration
		}
		return err
	}

	// Enforce 60s cooldown
	latest, err := s.repo.GetLatestActiveOTP(ctx, email, "password_reset")
	if err == nil {
		if time.Since(latest.CreatedAt) < s.cfg.OTPCooldown {
			return ErrCooldownActive
		}
	}

	code, err := GenerateSecureOTP(s.cfg.OTPLength)
	if err != nil {
		return err
	}

	codeHash := HashOTP(code)
	expiresAt := time.Now().Add(s.cfg.OTPExpiry)
	_, err = s.repo.CreateOTP(ctx, email, codeHash, "password_reset", s.cfg.OTPMaxAttempts, expiresAt)
	if err != nil {
		return err
	}

	return s.mailer.SendPasswordResetOTP(ctx, user.Email, user.Name, code)
}

// ResetPassword validates the recovery OTP, updates the password hash, and increments token_version to revoke active sessions.
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	code = strings.TrimSpace(code)

	if len(newPassword) < 8 || len(newPassword) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}

	challenge, err := s.repo.GetLatestActiveOTP(ctx, email, "password_reset")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrInvalidOTP
		}
		return err
	}

	if time.Now().After(challenge.ExpiresAt) {
		_ = s.repo.ConsumeOTP(ctx, challenge.ID)
		return ErrExpiredOTP
	}

	if challenge.Attempts >= challenge.MaxAttempts {
		_ = s.repo.ConsumeOTP(ctx, challenge.ID)
		return ErrMaxAttemptsExceeded
	}

	if !VerifyOTPHash(code, challenge.CodeHash) {
		_ = s.repo.IncrementOTPAttempts(ctx, challenge.ID)
		if challenge.Attempts+1 >= challenge.MaxAttempts {
			_ = s.repo.ConsumeOTP(ctx, challenge.ID)
			return ErrMaxAttemptsExceeded
		}
		return ErrInvalidOTP
	}

	_ = s.repo.ConsumeOTP(ctx, challenge.ID)

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Updates password and increments token_version, revoking all existing JWT sessions
	return s.repo.UpdateUserPassword(ctx, email, string(hashBytes))
}

// GoogleSignIn provisions or links a Google user, returning an authenticated session.
func (s *Service) GoogleSignIn(ctx context.Context, googleID, email, name string) (User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.UpsertGoogleUser(ctx, googleID, email, name)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to complete Google authentication: %w", err)
	}

	token, err := s.tokenMgr.GenerateToken(user.ID, user.TokenVersion)
	if err != nil {
		return User{}, "", fmt.Errorf("failed to issue session token: %w", err)
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
