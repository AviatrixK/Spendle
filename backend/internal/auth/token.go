package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager manages creation and cryptographic verification of JWT authentication tokens.
type TokenManager struct {
	secret []byte
}

// NewTokenManager initializes a new TokenManager with the provided secret key.
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

// GenerateToken produces an HS256 signed JWT with `sub`, `iat`, and a 7-day expiration.
func (tm *TokenManager) GenerateToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

// ValidateToken parses and cryptographically validates the token string.
//
// Go Concept - JWT Algorithm Confusion Defense:
// A common vulnerability in JWT implementations is accepting an attacker's chosen header algorithm.
// For example, if an attacker changes the header to `"alg": "none"`, insecure libraries might treat
// the token as valid without verifying any signature.
// By strictly verifying `t.Method.(*jwt.SigningMethodHMAC)`, we guarantee that tokens signed with
// `"none"`, asymmetric keys, or unsupported algorithms are rejected immediately.
func (tm *TokenManager) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v (HMAC required)", t.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}

