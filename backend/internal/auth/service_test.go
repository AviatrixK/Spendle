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
	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/config"
)

const testSecret = "super-secret-key-at-least-32-characters-long!!"

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:      testSecret,
		OTPExpiry:      2 * time.Minute,
		OTPCooldown:    60 * time.Second,
		OTPMaxAttempts: 3,
		OTPLength:      4,
	}
}

// fakeMailer records sent emails and OTPs in memory for test assertions.
type fakeMailer struct {
	mu            sync.Mutex
	lastEmail     string
	lastName      string
	lastOTP       string
	lastPurpose   string
	sendErr       error
	dispatchCount int
}

func (m *fakeMailer) SendVerificationOTP(ctx context.Context, toEmail, toName, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.lastEmail = toEmail
	m.lastName = toName
	m.lastOTP = otp
	m.lastPurpose = "email_verification"
	m.dispatchCount++
	return nil
}

func (m *fakeMailer) SendPasswordResetOTP(ctx context.Context, toEmail, toName, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.lastEmail = toEmail
	m.lastName = toName
	m.lastOTP = otp
	m.lastPurpose = "password_reset"
	m.dispatchCount++
	return nil
}

// fakeRepository is a thread-safe in-memory implementation of auth.Repository.
type fakeRepository struct {
	mu      sync.RWMutex
	users   map[string]auth.User
	byEmail map[string]string
	otps    map[string]auth.OTPChallenge
	nextID  int
	otpID   int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		users:   make(map[string]auth.User),
		byEmail: make(map[string]string),
		otps:    make(map[string]auth.OTPChallenge),
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
		PasswordHash: &passwordHash,
		IsVerified:   false,
		TokenVersion: 1,
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

func (r *fakeRepository) VerifyUserEmail(ctx context.Context, email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedEmail := strings.ToLower(email)
	id, exists := r.byEmail[normalizedEmail]
	if !exists {
		return auth.ErrNotFound
	}

	u := r.users[id]
	now := time.Now()
	u.IsVerified = true
	u.VerifiedAt = &now
	r.users[id] = u
	return nil
}

func (r *fakeRepository) UpdateUserPassword(ctx context.Context, email, newPasswordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedEmail := strings.ToLower(email)
	id, exists := r.byEmail[normalizedEmail]
	if !exists {
		return auth.ErrNotFound
	}

	u := r.users[id]
	u.PasswordHash = &newPasswordHash
	u.TokenVersion++
	r.users[id] = u
	return nil
}

func (r *fakeRepository) UpsertGoogleUser(ctx context.Context, googleID, email, name string) (auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedEmail := strings.ToLower(email)
	if id, exists := r.byEmail[normalizedEmail]; exists {
		u := r.users[id]
		u.GoogleID = &googleID
		u.IsVerified = true
		r.users[id] = u
		return u, nil
	}

	r.nextID++
	id := fmt.Sprintf("user-%d", r.nextID)
	now := time.Now()
	u := auth.User{
		ID:           id,
		Name:         name,
		Email:        normalizedEmail,
		GoogleID:     &googleID,
		IsVerified:   true,
		VerifiedAt:   &now,
		TokenVersion: 1,
		CreatedAt:    now,
	}
	r.users[id] = u
	r.byEmail[normalizedEmail] = id
	return u, nil
}

func (r *fakeRepository) CreateOTP(ctx context.Context, email, codeHash, purpose string, maxAttempts int, expiresAt time.Time) (auth.OTPChallenge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	normalizedEmail := strings.ToLower(email)
	now := time.Now()
	// Invalidate previous challenges
	for k, o := range r.otps {
		if o.Email == normalizedEmail && o.Purpose == purpose && o.ConsumedAt == nil {
			o.ConsumedAt = &now
			r.otps[k] = o
		}
	}

	r.otpID++
	id := fmt.Sprintf("otp-%d", r.otpID)
	challenge := auth.OTPChallenge{
		ID:          id,
		Email:       normalizedEmail,
		CodeHash:    codeHash,
		Purpose:     purpose,
		Attempts:    0,
		MaxAttempts: maxAttempts,
		ExpiresAt:   expiresAt,
		ConsumedAt:  nil,
		CreatedAt:   now,
	}
	r.otps[id] = challenge
	return challenge, nil
}

func (r *fakeRepository) GetLatestActiveOTP(ctx context.Context, email, purpose string) (auth.OTPChallenge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedEmail := strings.ToLower(email)
	var latest auth.OTPChallenge
	var found bool

	for _, o := range r.otps {
		if o.Email == normalizedEmail && o.Purpose == purpose && o.ConsumedAt == nil {
			if !found || o.CreatedAt.After(latest.CreatedAt) {
				latest = o
				found = true
			}
		}
	}

	if !found {
		return auth.OTPChallenge{}, auth.ErrNotFound
	}
	return latest, nil
}

func (r *fakeRepository) IncrementOTPAttempts(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, exists := r.otps[id]
	if !exists {
		return auth.ErrNotFound
	}
	o.Attempts++
	r.otps[id] = o
	return nil
}

func (r *fakeRepository) ConsumeOTP(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, exists := r.otps[id]
	if !exists {
		return auth.ErrNotFound
	}
	now := time.Now()
	o.ConsumedAt = &now
	r.otps[id] = o
	return nil
}

// 1. Test Signup creates unverified account and dispatches 4-digit OTP
func TestSignup_SuccessAndOTPDispatch(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	mailer := &fakeMailer{}
	cfg := testConfig()
	svc := auth.NewService(repo, tokenMgr, mailer, cfg)

	ctx := context.Background()
	user, err := svc.Signup(ctx, "  Jane Doe  ", "Jane.Doe@EXAMPLE.COM", "password123")
	if err != nil {
		t.Fatalf("unexpected error during signup: %v", err)
	}

	if user.Email != "jane.doe@example.com" {
		t.Errorf("expected normalized email 'jane.doe@example.com', got %q", user.Email)
	}
	if user.Name != "Jane Doe" {
		t.Errorf("expected trimmed name 'Jane Doe', got %q", user.Name)
	}
	if user.IsVerified {
		t.Error("newly registered account should NOT be marked verified before OTP confirmation")
	}

	if mailer.dispatchCount != 1 {
		t.Fatalf("expected 1 email dispatch, got %d", mailer.dispatchCount)
	}
	if len(mailer.lastOTP) != 4 {
		t.Errorf("expected 4-digit OTP, got %q", mailer.lastOTP)
	}
}

// 2. Test Verification with correct OTP marks verified and returns session token
func TestVerifyEmail_Success(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	mailer := &fakeMailer{}
	cfg := testConfig()
	svc := auth.NewService(repo, tokenMgr, mailer, cfg)
	ctx := context.Background()

	_, err := svc.Signup(ctx, "Jane Doe", "jane@example.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	otpCode := mailer.lastOTP
	user, token, err := svc.VerifyEmail(ctx, "jane@example.com", otpCode)
	if err != nil {
		t.Fatalf("email verification failed: %v", err)
	}

	if !user.IsVerified {
		t.Error("expected user.IsVerified to be true after successful verification")
	}
	if token == "" {
		t.Error("expected non-empty session token after verification")
	}

	// Verify token validity
	subject, tokenVer, err := tokenMgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if subject != user.ID {
		t.Errorf("expected subject %s, got %s", user.ID, subject)
	}
	if tokenVer != user.TokenVersion {
		t.Errorf("expected tokenVersion %d, got %d", user.TokenVersion, tokenVer)
	}
}

// 3. Test Verification with wrong OTP increments attempts and rejects
func TestVerifyEmail_InvalidOTPAndMaxAttempts(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	mailer := &fakeMailer{}
	cfg := testConfig()
	svc := auth.NewService(repo, tokenMgr, mailer, cfg)
	ctx := context.Background()

	_, err := svc.Signup(ctx, "Jane Doe", "jane@example.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	// Attempt 1: wrong OTP
	_, _, err = svc.VerifyEmail(ctx, "jane@example.com", "0000")
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP on attempt 1, got %v", err)
	}

	// Attempt 2: wrong OTP
	_, _, err = svc.VerifyEmail(ctx, "jane@example.com", "1111")
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP on attempt 2, got %v", err)
	}

	// Attempt 3: wrong OTP reaches max attempts ceiling
	_, _, err = svc.VerifyEmail(ctx, "jane@example.com", "2222")
	if !errors.Is(err, auth.ErrMaxAttemptsExceeded) {
		t.Fatalf("expected ErrMaxAttemptsExceeded on attempt 3, got %v", err)
	}

	// Attempt 4: challenge is now consumed
	_, _, err = svc.VerifyEmail(ctx, "jane@example.com", mailer.lastOTP)
	if !errors.Is(err, auth.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP after challenge consumed, got %v", err)
	}
}

// 4. Test Login blocks unverified accounts with ErrEmailUnverified
func TestLogin_UnverifiedAccountBlocked(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	mailer := &fakeMailer{}
	cfg := testConfig()
	svc := auth.NewService(repo, tokenMgr, mailer, cfg)
	ctx := context.Background()

	_, err := svc.Signup(ctx, "Jane Doe", "jane@example.com", "password123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}

	// Login before email verification
	_, _, err = svc.Login(ctx, "jane@example.com", "password123")
	if !errors.Is(err, auth.ErrEmailUnverified) {
		t.Fatalf("expected ErrEmailUnverified before verification, got: %v", err)
	}

	// Verify email
	_, _, err = svc.VerifyEmail(ctx, "jane@example.com", mailer.lastOTP)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}

	// Login after email verification succeeds
	user, token, err := svc.Login(ctx, "jane@example.com", "password123")
	if err != nil {
		t.Fatalf("login should succeed after verification, got: %v", err)
	}
	if token == "" || user.ID == "" {
		t.Error("expected valid user and token on successful login")
	}
}

// 5. Test Password Reset flow and Token Versioning Revocation
func TestPasswordReset_RevokesExistingSessions(t *testing.T) {
	repo := newFakeRepository()
	tokenMgr := auth.NewTokenManager(testSecret)
	mailer := &fakeMailer{}
	cfg := testConfig()
	svc := auth.NewService(repo, tokenMgr, mailer, cfg)
	ctx := context.Background()

	// Register and verify
	_, err := svc.Signup(ctx, "Alice", "alice@example.com", "oldPassword123")
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	_, oldToken, err := svc.VerifyEmail(ctx, "alice@example.com", mailer.lastOTP)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}

	// Validate that old token is initially valid
	_, initialVersion, err := tokenMgr.ValidateToken(oldToken)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}

	// Request password reset
	err = svc.ForgotPassword(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("forgot password failed: %v", err)
	}
	resetOTP := mailer.lastOTP

	// Reset password
	err = svc.ResetPassword(ctx, "alice@example.com", resetOTP, "newSuperPassword456")
	if err != nil {
		t.Fatalf("reset password failed: %v", err)
	}

	// Verify user's database token_version incremented
	user, err := repo.GetUserByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("failed to retrieve user: %v", err)
	}
	if user.TokenVersion <= initialVersion {
		t.Errorf("expected token_version to increase (> %d), got %d", initialVersion, user.TokenVersion)
	}

	// Login with old password fails
	_, _, err = svc.Login(ctx, "alice@example.com", "oldPassword123")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials with old password, got %v", err)
	}

	// Login with new password succeeds
	_, newToken, err := svc.Login(ctx, "alice@example.com", "newSuperPassword456")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	_, newClaimVersion, err := tokenMgr.ValidateToken(newToken)
	if err != nil {
		t.Fatalf("failed to validate new token: %v", err)
	}
	if newClaimVersion != user.TokenVersion {
		t.Errorf("expected new token claim version %d, got %d", user.TokenVersion, newClaimVersion)
	}
}

// 6. Test Token Validation Rejections (Different Secret, Expired, and 'none' Algorithm)
func TestToken_ValidationRejections(t *testing.T) {
	tokenMgr := auth.NewTokenManager(testSecret)

	t.Run("token signed with different secret is rejected", func(t *testing.T) {
		foreignMgr := auth.NewTokenManager("different-secret-key-at-least-32-chars-long!!")
		foreignToken, err := foreignMgr.GenerateToken("user-foreign", 1)
		if err != nil {
			t.Fatalf("failed to generate foreign token: %v", err)
		}

		_, _, err = tokenMgr.ValidateToken(foreignToken)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for foreign signature, got: %v", err)
		}
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		expiredClaims := auth.SpendleClaims{
			TokenVersion: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-expired",
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		}
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		signed, err := expiredToken.SignedString([]byte(testSecret))
		if err != nil {
			t.Fatalf("failed to create expired token: %v", err)
		}

		_, _, err = tokenMgr.ValidateToken(signed)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for expired token, got: %v", err)
		}
	})

	t.Run("token with 'none' algorithm is rejected", func(t *testing.T) {
		noneClaims := auth.SpendleClaims{
			TokenVersion: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-hacker",
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, noneClaims)
		signedNone, err := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("failed to create 'none' token: %v", err)
		}

		_, _, err = tokenMgr.ValidateToken(signedNone)
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken for 'none' algorithm token, got: %v", err)
		}
	})
}
