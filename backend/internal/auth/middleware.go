package auth

import (
	"context"
	"net/http"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// Go Concept - Typed Context Keys:
// Using primitive types (like string or int) for context keys can lead to collisions across packages.
// Go best practice is to declare an unexported struct type (`type contextKey struct{}`) and use it
// as the key. Because the type is unexported, no other package can overwrite or forge this key.
type contextKey struct{}

var userIDContextKey = contextKey{}

// UserIDFromContext extracts the authenticated user ID from the request context.
// Returns the user ID string and true if present, or ("", false) if unauthenticated.
func UserIDFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok
}

// RequireAuth reads the session cookie, validates the JWT, checks that the user's
// database token_version has not been revoked (e.g. after a password change), and loads the user ID into context.
func RequireAuth(tokenMgr *TokenManager, repo Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(CookieName)
			if err != nil || cookie.Value == "" {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			userID, claimVersion, err := tokenMgr.ValidateToken(cookie.Value)
			if err != nil {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired authentication token")
				return
			}

			// Go Concept - Instant Session Revocation via Token Versioning:
			// If a user changed their password or triggered a security logout, their database `token_version`
			// is incremented. Any older JWT in circulation carrying an outdated version is rejected immediately.
			user, err := repo.GetUserByID(r.Context(), userID)
			if err != nil || user.TokenVersion != claimVersion {
				_ = httpx.WriteError(w, http.StatusUnauthorized, "session has been invalidated; please log in again")
				return
			}

			// Attach the validated user ID to the request context
			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
