package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/server"
)

func main() {
	// 1. Load configuration from environment variables (and optional .env file)
	cfg, err := config.Load()
	if err != nil {
		// Go Concept - slog (Structured Logging):
		// Go 1.21+ has built-in structured logging via `log/slog`.
		// It supports key-value pairs instead of unformatted print strings.
		slog.Error("configuration loading failed", "error", err)
		os.Exit(1)
	}

	// In development, text logs are easy to read; in production, JSON logs are easily ingested.
	var logHandler slog.Handler
	if cfg.Env == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(logHandler)

	// 2. Setup graceful shutdown notification context
	// Go Concept - signal.NotifyContext:
	// signal.NotifyContext creates a Context that gets marked as Done when the OS delivers
	// an interrupt signal (SIGINT / Ctrl+C or SIGTERM from process managers/Docker).
	// 'defer stop()' restores default signal behavior when main exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Initialize server with configuration and logger
	srv := server.New(cfg, logger)

	// 4. Start the HTTP server inside a background goroutine
	// Go Concept - Goroutines:
	// A goroutine is a lightweight thread of execution managed by the Go runtime (prefix with `go`).
	// Because `srv.Start()` is blocking (it loops forever accepting TCP connections), running it
	// in a goroutine allows the main function to continue and listen for shutdown signals.
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server encountered fatal error", "error", err)
			os.Exit(1)
		}
	}()

	// 5. Block the main thread until an OS signal is caught
	// Go Concept - Channel Receive (<-):
	// `ctx.Done()` returns a channel that closes when a signal arrives. Reading from it blocks
	// the main goroutine without consuming CPU cycles until the event occurs.
	<-ctx.Done()

	logger.Info("shutdown signal received, initiating graceful shutdown")

	// 6. Create a 10-second timeout context for in-flight requests to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown due to timeout", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited cleanly")
}

