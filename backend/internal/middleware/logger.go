package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the HTTP status code.
// Go Concept - Struct Embedding:
// By embedding `http.ResponseWriter` directly inside our struct, our struct automatically inherits
// all its methods (like Header(), Write()). We only need to override WriteHeader() to intercept the code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	// Default status is 200 OK unless WriteHeader is explicitly called
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Go Concept - HTTP Middleware Pattern:
// In Go, middleware is a higher-order function that takes an `http.Handler` and returns a new `http.Handler`.
// It runs logic before calling `next.ServeHTTP(w, r)`, similar to `(req, res, next) => {}` in Express.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriter(w)

			// Pass control to the next handler in the chain
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration.String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

