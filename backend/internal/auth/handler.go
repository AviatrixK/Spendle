package auth

import (
	"errors"
	"net/http"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// CookieName is the key used for the authentication JWT stored in the user's browser.
const CookieName = "spendle_token"

// Handler handles HTTP requests for user registration, authentication, and sessions.
type Handler struct {
	service      *Service
	isProduction bool
}

// NewHandler constructs an auth HTTP Handler.
func NewHandler(service *Service, isProduction bool) *Handler {
	return &Handler{
		service:      service,
		isProduction: isProduction,
	}
}

// SignupRequest defines the JSON payload for user registration.
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest defines the JSON payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup handles POST /api/v1/auth/signup.
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.service.Signup(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			_ = httpx.WriteError(w, http.StatusConflict, err.Error())
			return
		}
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.setAuthCookie(w, token)
	_ = httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"user": user.ToResponse(),
	})
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
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

// Go Concept - Cookie Security Flags:
// - HttpOnly: true -> Disallows document.cookie access from client JavaScript, completely blocking XSS token theft.
// - Path: "/" -> Ensures the browser automatically attaches the cookie to all backend routes.
// - SameSite: Lax -> Prevents CSRF attacks by forbidding the browser from sending the cookie on third-party requests.
// - Secure: true in production -> Enforces HTTPS transmission only (disabled on localhost HTTP during development).
// - MaxAge: 7 * 24 * 3600 (604,800s) -> Sets cookie lifespan to match token validity (7 days).
func (h *Handler) setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookie instructs the client's browser to delete the cookie immediately.
func (h *Handler) clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Instructs browser to expire immediately
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})
}

