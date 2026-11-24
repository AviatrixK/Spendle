package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SpendleClaims represents the structured claims encoded within the authentication JWT.
type SpendleClaims struct {
	jwt.RegisteredClaims
	TokenVersion int `json:"ver"`
}

// TokenManager manages creation and cryptographic verification of JWT authentication tokens.
type TokenManager struct {
	secret []byte
}

// NewTokenManager initializes a new TokenManager with the provided secret key.
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

// GenerateToken produces an HS256 signed JWT with `sub`, `ver`, `iat`, and a 7-day expiration.
func (tm *TokenManager) GenerateToken(userID string, tokenVersion int) (string, error) {
	now := time.Now()
	claims := SpendleClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
		TokenVersion: tokenVersion,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

// ValidateToken parses and cryptographically validates the token string, asserting the HMAC algorithm.
// Returns the subject user ID and token version encoded in the claims.
func (tm *TokenManager) ValidateToken(tokenStr string) (string, int, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SpendleClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v (HMAC required)", t.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil || !token.Valid {
		return "", 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*SpendleClaims)
	if !ok || claims.Subject == "" {
		return "", 0, ErrInvalidToken
	}

	return claims.Subject, claims.TokenVersion, nil
}
