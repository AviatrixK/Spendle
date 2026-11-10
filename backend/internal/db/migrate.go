package db

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Go Concept - //go:embed:
// The `//go:embed` directive tells the Go compiler at build time to read files from disk
// and embed their raw bytes directly into the compiled binary.
// This means our database migration SQL files become part of the executable, so we never have
// to worry about missing SQL files when shipping the binary into Docker or cloud containers.

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations executes all pending database migrations against the database.
// It uses golang-migrate with the `iofs` source driver to read embedded SQL files.
// If all migrations have already been applied, it treats `migrate.ErrNoChange` as success.
func RunMigrations(databaseURL string) error {
	// golang-migrate's pgx/v5 driver is registered under the "pgx5" URL scheme.
	// We normalize standard "postgres://" and "postgresql://" URLs so that callers can
	// pass their standard connection string without needing to know the internal scheme name.
	migrateURL := databaseURL
	if strings.HasPrefix(migrateURL, "postgres://") {
		migrateURL = "pgx5://" + strings.TrimPrefix(migrateURL, "postgres://")
	} else if strings.HasPrefix(migrateURL, "postgresql://") {
		migrateURL = "pgx5://" + strings.TrimPrefix(migrateURL, "postgresql://")
	}

	// Go Concept - iofs (io/fs Driver):
	// The standard library `embed.FS` implements Go's `io/fs.FS` interface.
	// The `iofs` driver bridges embedded filesystems into `golang-migrate`, allowing migrations
	// to be discovered and executed entirely in-memory.
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration file driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", driver, migrateURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migration instance: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		_ = srcErr
		_ = dbErr
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

