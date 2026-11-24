package auth

import "errors"

// Sentinel errors for authentication, OTP challenges, and security checks.
var (
	// ErrEmailTaken is returned when registration is attempted with an already existing email address.
	ErrEmailTaken = errors.New("email is already registered")

	// ErrNotFound is returned when a requested user does not exist in the database.
	ErrNotFound = errors.New("user not found")

	// ErrInvalidCredentials is returned for both non-existent emails and incorrect passwords
	// to prevent user enumeration attacks.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrInvalidToken is returned when an authentication JWT is missing, malformed, expired, or revoked.
	ErrInvalidToken = errors.New("invalid or expired authentication token")

	// ErrInvalidOTP is returned when the provided 4-digit code does not match the stored cryptographic hash.
	ErrInvalidOTP = errors.New("invalid or incorrect verification code")

	// ErrExpiredOTP is returned when attempting to verify a code past its validity window (2 minutes).
	ErrExpiredOTP = errors.New("verification code has expired; please request a new code")

	// ErrMaxAttemptsExceeded is returned when 3 failed attempts are reached, invalidating the challenge.
	ErrMaxAttemptsExceeded = errors.New("maximum verification attempts exceeded; please request a new code")

	// ErrCooldownActive is returned when a user requests an OTP resend before the 60s cooldown expires.
	ErrCooldownActive = errors.New("please wait before requesting another verification code")

	// ErrEmailUnverified is returned when an unverified account attempts to log in.
	ErrEmailUnverified = errors.New("email_unverified")

	// ErrPasswordMismatch is returned when password and confirm_password do not match.
	ErrPasswordMismatch = errors.New("passwords do not match")
)
