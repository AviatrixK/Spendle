package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// CookieName is the key used for the authentication JWT stored in the user's browser.
const CookieName = "spendle_token"

// GoogleOAuthStateCookie is the cookie key used to store the CSRF state during Google OAuth login.
const GoogleOAuthStateCookie = "spendle_oauth_state"

// Handler handles HTTP requests for user registration, authentication, sessions, OTPs, and OAuth.
type Handler struct {
	service *Service
	cfg     *config.Config
}

// NewHandler constructs an auth HTTP Handler.
func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

// SignupRequest defines the JSON payload for initial user registration.
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyEmailRequest defines the JSON payload for verifying an email via 4-digit OTP.
type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// ResendOTPRequest defines the JSON payload for re-dispatching an OTP.
type ResendOTPRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

// LoginRequest defines the JSON payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ForgotPasswordRequest defines the JSON payload for initiating password recovery.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest defines the JSON payload for resetting password via OTP.
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// Signup handles POST /api/v1/auth/signup.
// Validates registration fields, creates an unverified account, and sends a 4-digit OTP to the user's email.
// Note: Does NOT issue a session cookie yet, since the email is unverified.
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Signup(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			_ = httpx.WriteError(w, http.StatusConflict, "an account with this email address already exists")
			return
		}
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"message": "Verification code sent to your email",
		"email":   user.Email,
	})
}

// VerifyEmail handles POST /api/v1/auth/verify-email.
// Validates the 4-digit OTP, marks the account as verified, issues a JWT session cookie, and returns user details.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" || req.Code == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "email and verification code are required")
		return
	}

	user, token, err := h.service.VerifyEmail(r.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidOTP):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid verification code")
		case errors.Is(err, ErrExpiredOTP):
			_ = httpx.WriteError(w, http.StatusBadRequest, "verification code has expired; please request a new one")
		case errors.Is(err, ErrMaxAttemptsExceeded):
			_ = httpx.WriteError(w, http.StatusTooManyRequests, "maximum verification attempts exceeded; please request a new code")
		default:
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.setAuthCookie(w, token)
	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Email verified successfully",
		"user":    user.ToResponse(),
	})
}

// ResendOTP handles POST /api/v1/auth/resend-otp.
// Invalidates active OTPs and sends a fresh 4-digit OTP subject to a 60-second cooldown.
func (h *Handler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var req ResendOTPRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = "email_verification"
	}

	if err := h.service.ResendOTP(r.Context(), req.Email, purpose); err != nil {
		if errors.Is(err, ErrCooldownActive) {
			_ = httpx.WriteError(w, http.StatusTooManyRequests, "please wait 60 seconds before requesting a new code")
			return
		}
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "A new verification code has been sent to your email",
	})
}

// Login handles POST /api/v1/auth/login.
// Authenticates credentials and verifies email status. If unverified, returns 403 Forbidden with email_unverified.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailUnverified) {
			// Trigger a fresh verification code so user can complete registration immediately
			_ = h.service.ResendOTP(r.Context(), user.Email, "email_verification")
			_ = httpx.WriteJSON(w, http.StatusForbidden, map[string]any{
				"error":   "email_unverified",
				"message": "Your email address is not verified. A new verification code has been sent to your inbox.",
				"email":   user.Email,
			})
			return
		}
		if errors.Is(err, ErrInvalidCredentials) {
			_ = httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.setAuthCookie(w, token)
	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user": user.ToResponse(),
	})
}

// ForgotPassword handles POST /api/v1/auth/forgot-password.
// Dispatches a 4-digit recovery OTP if the account exists, returning an opaque message to prevent enumeration.
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}

	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		if errors.Is(err, ErrCooldownActive) {
			_ = httpx.WriteError(w, http.StatusTooManyRequests, "please wait 60 seconds before requesting a new reset code")
			return
		}
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "If an account exists with this email address, a password reset code has been sent.",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password.
// Validates the recovery OTP, updates the password, and invalidates all active sessions via token_version.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" || req.Code == "" || req.NewPassword == "" {
		_ = httpx.WriteError(w, http.StatusBadRequest, "email, code, and new_password are required")
		return
	}

	if err := h.service.ResetPassword(r.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, ErrInvalidOTP):
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid reset code")
		case errors.Is(err, ErrExpiredOTP):
			_ = httpx.WriteError(w, http.StatusBadRequest, "reset code has expired; please request a new one")
		case errors.Is(err, ErrMaxAttemptsExceeded):
			_ = httpx.WriteError(w, http.StatusTooManyRequests, "maximum attempts exceeded; please request a new reset code")
		default:
			_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully. You can now log in with your new password.",
	})
}

// GoogleLogin handles GET /api/v1/auth/google/login.
// Initiates the Google OAuth 2.0 flow if configured; returns 501 Not Implemented if unconfigured.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.cfg.GoogleClientID == "" || h.cfg.GoogleClientSecret == "" {
		_ = httpx.WriteError(w, http.StatusNotImplemented, "Google sign-in is not configured on this server")
		return
	}

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to generate authentication state")
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Set short-lived state cookie (5 minutes) for CSRF validation
	http.SetCookie(w, &http.Cookie{
		Name:     GoogleOAuthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})

	googleAuthURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&prompt=select_account",
		url.QueryEscape(h.cfg.GoogleClientID),
		url.QueryEscape(h.cfg.GoogleRedirectURL),
		url.QueryEscape("openid email profile"),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, googleAuthURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /api/v1/auth/google/callback.
// Validates state, exchanges the authorization code for an ID token/user profile, sets session cookie,
// and redirects to the frontend dashboard.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.cfg.GoogleClientID == "" || h.cfg.GoogleClientSecret == "" {
		_ = httpx.WriteError(w, http.StatusNotImplemented, "Google sign-in is not configured on this server")
		return
	}

	// Verify CSRF state cookie
	stateCookie, err := r.Cookie(GoogleOAuthStateCookie)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=invalid_oauth_state", http.StatusTemporaryRedirect)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     GoogleOAuthStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=missing_auth_code", http.StatusTemporaryRedirect)
		return
	}

	// Exchange authorization code for access token
	tokenEndpoint := "https://oauth2.googleapis.com/token"
	formValues := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.GoogleClientID},
		"client_secret": {h.cfg.GoogleClientSecret},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"grant_type":    {"authorization_code"},
	}

	reqCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tokenReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, tokenEndpoint, strings.NewReader(formValues.Encode()))
	if err != nil {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=oauth_exchange_failed", http.StatusTemporaryRedirect)
		return
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(tokenReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=oauth_token_error", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil || tokenData.AccessToken == "" {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=oauth_decode_error", http.StatusTemporaryRedirect)
		return
	}

	// Fetch user profile from Google UserInfo endpoint
	userinfoReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=userinfo_request_failed", http.StatusTemporaryRedirect)
		return
	}
	userinfoReq.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

	userinfoResp, err := client.Do(userinfoReq)
	if err != nil || userinfoResp.StatusCode != http.StatusOK {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=userinfo_fetch_failed", http.StatusTemporaryRedirect)
		return
	}
	defer userinfoResp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(userinfoResp.Body).Decode(&googleUser); err != nil || googleUser.Email == "" {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error=userinfo_parse_error", http.StatusTemporaryRedirect)
		return
	}

	_, sessionToken, err := h.service.GoogleSignIn(r.Context(), googleUser.ID, googleUser.Email, googleUser.Name)
	if err != nil {
		http.Redirect(w, r, h.cfg.FrontendOrigin+"/login?error="+url.QueryEscape(err.Error()), http.StatusTemporaryRedirect)
		return
	}

	h.setAuthCookie(w, sessionToken)
	http.Redirect(w, r, h.cfg.FrontendOrigin+"/dashboard", http.StatusTemporaryRedirect)
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /api/v1/auth/me (protected by RequireAuth).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok || userID == "" {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.service.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		_ = httpx.WriteError(w, http.StatusInternalServerError, "failed to retrieve user profile")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user": user.ToResponse(),
	})
}

// setAuthCookie stores the JWT in an httpOnly secure cookie.
func (h *Handler) setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookie instructs the client browser to immediately expire the authentication cookie.
func (h *Handler) clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
	})
}
