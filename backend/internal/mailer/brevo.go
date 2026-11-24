package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const brevoAPIEndpoint = "https://api.brevo.com/v3/smtp/email"

// BrevoMailer implements Mailer using Brevo's v3 transactional email API.
type BrevoMailer struct {
	apiKey      string
	senderEmail string
	senderName  string
	httpClient  *http.Client
	logger      *slog.Logger
}

// NewBrevoMailer constructs a new Brevo transactional email client.
func NewBrevoMailer(apiKey, senderEmail, senderName string, logger *slog.Logger) *BrevoMailer {
	return &BrevoMailer{
		apiKey:      apiKey,
		senderEmail: senderEmail,
		senderName:  senderName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

type brevoRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoSender struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type brevoEmailPayload struct {
	Sender      brevoSender      `json:"sender"`
	To          []brevoRecipient `json:"to"`
	Subject     string           `json:"subject"`
	HTMLContent string           `json:"htmlContent"`
	TextContent string           `json:"textContent"`
}

// SendVerificationOTP sends the branded email verification OTP to a new user.
func (m *BrevoMailer) SendVerificationOTP(ctx context.Context, toEmail, toName, otp string) error {
	subject := fmt.Sprintf("Spendle — Your Verification Code: %s", otp)
	html := m.renderTemplate(
		"Verify Your Email",
		"Welcome to Spendle! Please enter the 4-digit code below to confirm your email and activate your account.",
		otp,
	)
	text := fmt.Sprintf("Welcome to Spendle!\n\nYour 4-digit verification code is: %s\n\nThis code expires in 2 minutes. Do not share it with anyone.", otp)

	return m.send(ctx, toEmail, toName, subject, html, text)
}

// SendPasswordResetOTP sends the branded password recovery OTP to the user.
func (m *BrevoMailer) SendPasswordResetOTP(ctx context.Context, toEmail, toName, otp string) error {
	subject := fmt.Sprintf("Spendle — Password Reset Code: %s", otp)
	html := m.renderTemplate(
		"Reset Your Password",
		"We received a request to reset your Spendle account password. Use the 4-digit recovery code below to proceed.",
		otp,
	)
	text := fmt.Sprintf("Spendle Password Reset\n\nYour 4-digit recovery code is: %s\n\nThis code expires in 2 minutes. If you did not request this, please ignore this email.", otp)

	return m.send(ctx, toEmail, toName, subject, html, text)
}

func (m *BrevoMailer) send(ctx context.Context, toEmail, toName, subject, htmlContent, textContent string) error {
	if m.apiKey == "" {
		m.logger.Warn("Brevo API key not configured; skipping outbound email dispatch", "to", toEmail)
		return errors.New("email delivery service is currently not configured")
	}

	payload := brevoEmailPayload{
		Sender: brevoSender{
			Email: m.senderEmail,
			Name:  m.senderName,
		},
		To: []brevoRecipient{
			{
				Email: toEmail,
				Name:  toName,
			},
		},
		Subject:     subject,
		HTMLContent: htmlContent,
		TextContent: textContent,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoAPIEndpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create Brevo HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", m.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Error("failed to connect to Brevo API", "error", err, "to", toEmail)
		return fmt.Errorf("network error communicating with email provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		m.logger.Error("Brevo API returned error status",
			"status_code", resp.StatusCode,
			"response", string(respBody),
			"to", toEmail,
		)
		return fmt.Errorf("email delivery failed with status code %d", resp.StatusCode)
	}

	m.logger.Info("transactional email dispatched successfully via Brevo", "to", toEmail, "subject", subject)
	return nil
}

// renderTemplate produces a modern, responsive HTML email with Spendle branding.
func (m *BrevoMailer) renderTemplate(title, description, otp string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin: 0; padding: 0; background-color: #0E0E11; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #EDEDEF;">
  <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #0E0E11; padding: 40px 16px;">
    <tr>
      <td align="center">
        <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 480px; background-color: #16161D; border: 1px solid rgba(255, 255, 255, 0.08); border-radius: 16px; overflow: hidden; box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);">
          <!-- Header -->
          <tr>
            <td style="padding: 32px 32px 16px 32px; text-align: center; border-bottom: 1px solid rgba(255, 255, 255, 0.05);">
              <h1 style="margin: 0; font-size: 28px; font-weight: 700; color: #FFFFFF; letter-spacing: -0.5px;">Spendle<span style="color: #E5A93C;">.</span></h1>
              <p style="margin: 6px 0 0 0; font-size: 11px; text-transform: uppercase; letter-spacing: 2px; color: #8E8E93;">Financial Discipline</p>
            </td>
          </tr>
          <!-- Content -->
          <tr>
            <td style="padding: 32px;">
              <h2 style="margin: 0 0 12px 0; font-size: 20px; font-weight: 600; color: #FFFFFF;">%s</h2>
              <p style="margin: 0 0 24px 0; font-size: 14px; line-height: 1.5; color: #A1A1A6;">%s</p>
              
              <!-- OTP Badge -->
              <div style="background: rgba(229, 169, 60, 0.08); border: 1px solid rgba(229, 169, 60, 0.3); border-radius: 12px; padding: 18px; text-align: center; margin-bottom: 24px;">
                <span style="font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace; font-size: 34px; font-weight: 700; letter-spacing: 12px; color: #E5A93C; margin-left: 12px;">%s</span>
              </div>

              <!-- Notice -->
              <p style="margin: 0 0 12px 0; font-size: 12px; color: #8E8E93; text-align: center;">
                ⏱️ Code valid for <strong>2 minutes</strong>. Maximum 3 attempts.
              </p>
              <p style="margin: 0; font-size: 12px; color: #636366; text-align: center;">
                If you did not request this code, you can safely disregard this email.
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding: 20px 32px; background-color: #121216; border-top: 1px solid rgba(255, 255, 255, 0.05); text-align: center;">
              <p style="margin: 0; font-size: 11px; color: #636366;">
                🔒 Secure verification powered by Brevo & Spendle Security Architecture.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, title, title, description, otp)
}

