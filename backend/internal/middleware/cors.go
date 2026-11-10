package middleware

import "net/http"

// Go Concept - Custom CORS Middleware:
// Cross-Origin Resource Sharing (CORS) is a browser security mechanism that restricts HTTP requests
// made from a different domain/port than the API server.
// Here we explicitly allow the configured frontend origin, allow credentials (cookies for JWT),
// and handle HTTP OPTIONS preflight requests.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If an origin is sent and matches our allowed frontend origin, grant access.
			// When credentials (cookies) are enabled, browsers disallow wildcard '*'.
			if origin == allowedOrigin || allowedOrigin == "*" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			}

			// Preflight request: browsers send an OPTIONS request before certain cross-origin calls
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

