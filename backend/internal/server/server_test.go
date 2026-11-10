package server_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/server"
)

// Go Concept - Testing in Go:
// Go includes a native testing framework in the `testing` package.
// Tests are written in files ending in `_test.go` with functions prefixed with `Test...`.
// `httptest.ResponseRecorder` records HTTP responses in memory without spinning up a real network socket.
func TestPingEndpoints(t *testing.T) {
	cfg := &config.Config{
		Port:           8080,
		Env:            "test",
		DatabaseURL:    "postgres://spendle:spendle@localhost:5432/spendle_db?sslmode=disable",
		JWTSecret:      "test-secret-must-be-at-least-32-chars-long!",
		FrontendOrigin: "http://localhost:5173",
	}
	// Discard log output during tests to keep test runs clean
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := server.New(cfg, logger, nil)

	endpoints := []string{"/ping", "/api/v1/ping"}

	for _, endpoint := range endpoints {
		t.Run("GET "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			rr := httptest.NewRecorder()

			srv.Router().ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
			}

			expectedContentType := "application/json"
			if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
				t.Errorf("expected Content-Type %q, got %q", expectedContentType, ct)
			}

			var body map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
				t.Fatalf("failed to parse JSON response body: %v", err)
			}

			if body["status"] != "ok" {
				t.Errorf("expected status 'ok', got %q", body["status"])
			}
			if body["app"] != "spendle" {
				t.Errorf("expected app 'spendle', got %q", body["app"])
			}
		})
	}
}

func TestHealthEndpoint_DatabaseUnavailable(t *testing.T) {
	cfg := &config.Config{
		Port:           8080,
		Env:            "test",
		DatabaseURL:    "postgres://spendle:spendle@localhost:5432/spendle_db?sslmode=disable",
		JWTSecret:      "test-secret-must-be-at-least-32-chars-long!",
		FrontendOrigin: "http://localhost:5173",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Pass nil db to simulate database unavailable
	srv := server.New(cfg, logger, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to parse JSON response body: %v", err)
	}

	if body["error"] != "database unavailable" {
		t.Errorf("expected error 'database unavailable', got %q", body["error"])
	}
}
