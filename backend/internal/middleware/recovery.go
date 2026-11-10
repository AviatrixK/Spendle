package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// Go Concept - Panic and Recover:
// In Go, a 'panic' is similar to an unhandled exception that unwinds the call stack and crashes the app.
// 'recover()' can only be called inside a 'defer' function. It stops the panicking sequence, captures
// the panic value, and lets the server recover gracefully instead of terminating the entire process.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					logger.Error("panic recovered",
						"error", fmt.Sprintf("%v", rec),
						"path", r.URL.Path,
						"method", r.Method,
						"stack", stack,
					)

					_ = httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

