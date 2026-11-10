package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Go Concept - Structs:
// A struct is a typed collection of fields, similar to an interface or class shape in TypeScript.
// Fields starting with a capital letter are "exported" (public to other packages).
type Config struct {
	Port           int
	Env            string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
}

// Go Concept - Pointers (*Config) and Multiple Return Values:
// Functions in Go often return `(result, error)`. Instead of throwing exceptions (try/catch),
// Go treats errors as ordinary values that must be checked explicitly.
// Returning `*Config` returns a pointer (memory address) to the struct to avoid copying it.
func Load() (*Config, error) {
	// Attempt to load .env file if present in the current working directory.
	// This keeps development convenient without adding external third-party dependencies.
	_ = loadDotEnv(".env")

	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value %q: %w", portStr, err)
	}

	env := getEnv("ENV", "development")
	dbURL := getEnv("DATABASE_URL", "")
	jwtSecret := getEnv("JWT_SECRET", "")
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:5173")

	cfg := &Config{
		Port:           port,
		Env:            env,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		FrontendOrigin: frontendOrigin,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Addr returns the network listen address string (e.g. ":8080").
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// validate checks constraints and ensures critical values are present.
func (c *Config) validate() error {
	switch c.Env {
	case "development", "production", "test":
		// valid environment
	default:
		return fmt.Errorf("invalid ENV %q: must be 'development', 'production', or 'test'", c.Env)
	}

	// DATABASE_URL is strictly required across all environments from Task 2 onwards
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required: environment variable is empty or not set")
	}

	// Fail-fast rule: in production, security secrets must never be omitted.
	if c.Env == "production" {
		if c.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required when ENV is 'production'")
		}
	}

	return nil
}

// getEnv retrieves an environment variable or falls back to a default value.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

// loadDotEnv is a lightweight parser that populates environment variables from a .env file.
// If the file does not exist, it silently returns nil. Existing env vars are not overwritten.
func loadDotEnv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err // file does not exist or cannot be read, ignore
	}
	// Go Concept - Defer:
	// defer schedules a function call to run right before the enclosing function returns.
	// This ensures resources (like file handles or database connections) are reliably cleaned up.
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty lines and comments
		}

		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// Strip wrapping quotes if present
		val = strings.Trim(val, `"'`)

		// Do not overwrite an already existing environment variable
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
