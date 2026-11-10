package auth_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/auth"
)

const testSecret = "super-secret-key-at-least-32-characters-long!!"

// fakeRepository is a thread-safe, in-memory implementation of auth.Repository for testing.
type fakeRepository struct {
	mu      sync.RWMutex
	users   map[string]auth.User
	byEmail map[string]string
	nextID  int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		users:   make(map[string]auth.User),
		byEmail: make(map[string]string),
	}
}

func (r *fakeRepository) CreateUser(ctx context.Context, name, email, passwordHash string) (auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedEmail := strings.ToLower(email)
	if _, exists := r.byEmail[normalizedEmail]; exists {
		return auth.User{}, auth.ErrEmailTaken
	}

	r.nextID++
	id := fmt.Sprintf("user-%d", r.nextID)

	user := auth.User{
		ID:           id,
		Name:         name,
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	r.users[id] = user
	r.byEmail[normalizedEmail] = id

	return user, nil
}

func (r *fakeRepository) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedEmail := strings.ToLower(email)
	id, exists := r.byEmail[normalizedEmail]
	if !exists {
		return auth.User{}, auth.ErrNotFound
	}
	return r.users[id], nil
}

func (r *fakeRepository) GetUserByID(ctx context.Context, id string) (auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return auth.User{}, auth.ErrNotFound
	}
	return user, nil
}

// 1. Test Signup Success and Email Normalisation
func TestSignup_SuccessAndNormalization(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	svc := auth.NewService(repo, tokenMgr)

	ctx := context.Background()
	user, token, err := svc.Signup(ctx, "  Jane Doe  ", "Jane.Doe@EXAMPLE.COM", "password123")
	if err != nil {
		t.Fatalf("unexpected error during signup: %v", err)
	}

	if user.Email != "jane.doe@example.com" {
		t.Errorf("expected normalized email 'jane.doe@example.com', got %q", user.Email)
	}
	if user.Name != "Jane Doe" {
		t.Errorf("expected trimmed name 'Jane Doe', got %q", user.Name)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify token resolves to created user ID
	subject, err := tokenMgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate issued token: %v", err)
	}
	if subject != user.ID {
		t.Errorf("expected token subject %q, got %q", user.ID, subject)
	}
}

// 2. Test Signup Rejects Invalid Inputs (Short Password, Invalid Email, Names)
func TestSignup_ValidationRejections(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	svc := auth.NewService(repo, tokenMgr)
	ctx := context.Background()

	tests := []struct {
		name     string
		userName string
		email    string
		password string
	}{
		{
			name:     "password too short (< 8 chars)",
			userName: "Valid User",
			email:    "valid@example.com",
			password: "short",
		},
		{
			name:     "password too long (> 72 chars)",
			userName: "Valid User",
			email:    "valid@example.com",
			password: strings.Repeat("a", 73),
		},
		{
			name:     "invalid email (missing @)",
			userName: "Valid User",
			email:    "invalidemail.com",
			password: "validPassword123",
		},
		{
			name:     "invalid email (missing domain dot)",
			userName: "Valid User",
			email:    "user@localhost",
			password: "validPassword123",
		},
		{
			name:     "name too short (< 2 chars)",
			userName: "A",
			email:    "valid@example.com",
			password: "validPassword123",
		},
		{
			name:     "name too long (> 60 chars)",
			userName: strings.Repeat("A", 61),
			email:    "valid@example.com",
			password: "validPassword123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Signup(ctx, tc.userName, tc.email, tc.password)
			if err == nil {
				t.Fatalf("expected error for case %q, but got nil", tc.name)
			}
		})
	}
}

// 3. Test Duplicate Email Returns ErrEmailTaken
func TestSignup_DuplicateEmail(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	svc := auth.NewService(repo, tokenMgr)
	ctx := context.Background()

	_, _, err := svc.Signup(ctx, "First User", "duplicate@example.com", "password123")
	if err != nil {
		t.Fatalf("initial signup failed: %v", err)
	}

	// Attempt duplicate signup with varied casing
	_, _, err = svc.Signup(ctx, "Second User", "DUPLICATE@example.com", "password456")
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

// 4. Test Login Success
func TestLogin_Success(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	svc := auth.NewService(repo, tokenMgr)
	ctx := context.Background()

	_, _, err := svc.Signup(ctx, "Login User", "login@example.com", "securePassword123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	// Login with uppercase email should succeed
	user, token, err := svc.Login(ctx, "LOGIN@EXAMPLE.COM", "securePassword123")
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}

	if user.Email != "login@example.com" {
		t.Errorf("expected email 'login@example.com', got %q", user.Email)
	}

	subject, err := tokenMgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate issued login token: %v", err)
	}
	if subject != user.ID {
		t.Errorf("expected subject %q, got %q", user.ID, subject)
	}
}

// 5. Test Login with Wrong Password and Unknown Email Both Return ErrInvalidCredentials
func TestLogin_SecurityRejections(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	svc := auth.NewService(repo, tokenMgr)
	ctx := context.Background()

	_, _, err := svc.Signup(ctx, "Alice", "alice@example.com", "correctPassword123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "unknown email",
			email:    "bob@example.com",
			password: "anyPassword123",
		},
		{
			name:     "known email with wrong password",
			email:    "alice@example.com",
			password: "wrongPassword!",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Login(ctx, tc.email, tc.password)
			if !errors.Is(err, auth.ErrInvalidCredentials) {
				t.Fatalf("expected ErrInvalidCredentials for case %q, got: %v", tc.name, err)
			}
		})
	}
}

// 6. Test Token Validation Rejections (Different Secret, Expired, and 'none' Algorithm)
func TestToken_ValidationRejections(t *testing.T) {
	tokenMgr := auth.NewTokenManager(testSecret)

	t.Run("token signed with different secret is rejected", func(t *testing.T) {
		foreignMgr := auth.NewTokenManager("different-secret-key-at-least-32-chars-long!!")
		foreignToken, err := foreignMgr.GenerateToken("user-foreign")
		if err != nil {
			t.Fatalf("failed to generate foreign token: %v", err)
		}

		_, err = tokenMgr.ValidateToken(foreignToken)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for foreign signature, got: %v", err)
		}
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		expiredClaims := jwt.RegisteredClaims{
			Subject:   "user-expired",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		}
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		signed, err := expiredToken.SignedString([]byte(testSecret))
		if err != nil {
			t.Fatalf("failed to create expired token: %v", err)
		}

		_, err = tokenMgr.ValidateToken(signed)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for expired token, got: %v", err)
		}
	})

	t.Run("token with 'none' algorithm is rejected", func(t *testing.T) {
		noneClaims := jwt.RegisteredClaims{
			Subject:   "user-hacker",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		}
		// Explicitly construct an unsigned token with alg: none
		noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, noneClaims)
		signedNone, err := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("failed to create 'none' token: %v", err)
		}

		_, err = tokenMgr.ValidateToken(signedNone)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for 'none' algorithm token, got: %v", err)
		}
	})
}

