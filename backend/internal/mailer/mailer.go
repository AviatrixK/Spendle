package mailer

import "context"

// Mailer defines the transactional email contract for authentication challenges.
// Go Concept - Interface Driven Design:
// By defining an interface, our business logic is decoupled from any specific third-party provider.
// During tests, an in-memory or mock mailer can be injected without making outbound network requests.
type Mailer interface {
	// SendVerificationOTP transmits a 4-digit email verification code during registration.
	SendVerificationOTP(ctx context.Context, toEmail, toName, otp string) error

	// SendPasswordResetOTP transmits a 4-digit code to authorize a password reset.
	SendPasswordResetOTP(ctx context.Context, toEmail, toName, otp string) error
}

