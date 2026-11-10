package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Go Concept - Database Connection Pooling (*pgxpool.Pool):
// Establishing a brand-new TCP socket and TLS handshake with PostgreSQL for every HTTP request
// is slow and resource-intensive. A connection pool maintains an active set of reusable connections.
// Concurrent HTTP requests (goroutines) borrow an open connection from the pool to execute queries
// and return it immediately when done, enabling high throughput without overloading the database.

// NewPool initializes and validates a PostgreSQL connection pool using pgxpool.
// It applies sensible production-ready defaults and verifies connectivity with a 5-second timeout.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	// Sensible pool configuration:
	// - MaxConns: maximum concurrent database connections allowed in the pool.
	// - MinConns: minimum idle connections kept open and ready for immediate work.
	// - MaxConnLifetime: maximum duration a connection is kept alive before being recycled.
	// - MaxConnIdleTime: duration an unused connection sits idle before being closed.
	// - HealthCheckPeriod: frequency at which idle connections are pinged to verify health.
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}

	// Go Concept - Context Timeouts:
	// `context.WithTimeout` produces a child context that automatically cancels after a duration
	// (here 5 seconds). If PostgreSQL is down or unreachable, the Ping operation will fail fast
	// instead of hanging the entire application startup indefinitely.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to database at %s:%d (ping failed or timed out): %w",
			poolConfig.ConnConfig.Host,
			poolConfig.ConnConfig.Port,
			err,
		)
	}

	return pool, nil
}

