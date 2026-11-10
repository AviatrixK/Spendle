package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/auth"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/middleware"
)

// Server encapsulates the HTTP server, routing engine, database connection pool,
// configuration, and structured logger.
// Go Concept - Composition over Inheritance:
// Go does not have OOP classes or inheritance. Instead, we compose structs by grouping
// fields together and declaring methods with a pointer receiver `(s *Server)`.
type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *pgxpool.Pool
	router *chi.Mux
	http   *http.Server
}

// New creates and configures a new Server instance.
// In Go, the idiom for constructors is a package function named `New(...)` or `NewServer(...)`.
func New(cfg *config.Config, logger *slog.Logger, db *pgxpool.Pool) *Server {
	router := chi.NewRouter()

	// Global Middlewares
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.Recoverer(logger))
	router.Use(middleware.CORS(cfg.FrontendOrigin))

	s := &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
		router: router,
		http: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	s.mountRoutes()

	return s
}

// mountRoutes registers all HTTP endpoints for the application.
func (s *Server) mountRoutes() {
	// Root-level ping handler
	pingHandler := func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"app":    "spendle",
		})
	}

	// Database connectivity health check handler
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		if s.db == nil {
			s.logger.Error("database health check failed: pool is nil")
			_ = httpx.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := s.db.Ping(ctx); err != nil {
			s.logger.Error("database health check failed", "error", err)
			_ = httpx.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}

		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":   "ok",
			"database": "up",
		})
	}

	// Root-level health check: GET /ping
	s.router.Get("/ping", pingHandler)

	// Dependency Injection: Initialize Auth components
	tokenMgr := auth.NewTokenManager(s.cfg.JWTSecret)
	authRepo := auth.NewPGRepository(s.db)
	authService := auth.NewService(authRepo, tokenMgr)
	authHandler := auth.NewHandler(authService, s.cfg.Env == "production")

	// API v1 group: mounts endpoints under /api/v1
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", pingHandler)
		r.Get("/health", healthHandler)

		// Auth group: /api/v1/auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", authHandler.Signup)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)

			// Protected routes requiring valid session cookie
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAuth(tokenMgr))
				r.Get("/me", authHandler.Me)
			})
		})
	})
}

// Start begins listening and serving incoming HTTP requests.
func (s *Server) Start() error {
	s.logger.Info("starting spendle api server",
		"addr", s.cfg.Addr(),
		"env", s.cfg.Env,
	)
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the server without interrupting active connections.
// Go Concept - Context (context.Context):
// Context carries deadlines, cancellation signals, and request-scoped values across API boundaries
// and goroutines. Here it enforces that the shutdown process aborts if it takes longer than the timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down spendle api server...")
	return s.http.Shutdown(ctx)
}

// Router returns the underlying Chi router (useful for tests using httptest).
func (s *Server) Router() *chi.Mux {
	return s.router
}
