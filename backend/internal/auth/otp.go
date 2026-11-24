package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateSecureOTP generates a cryptographically random numeric string preserving leading zeroes.
// Go Concept - crypto/rand vs math/rand:
// Standard `math/rand` produces pseudo-random numbers with predictable sequences if seeded poorly.
// `crypto/rand` queries operating system entropy sources (/dev/urandom on Unix, CryptGenRandom on Windows),
// making passcodes non-deterministic and cryptographically safe against prediction attacks.
func GenerateSecureOTP(digits int) (string, error) {
	if digits <= 0 {
		digits = 4
	}

	max := big.NewInt(1)
	for i := 0; i < digits; i++ {
		max.Mul(max, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate cryptographic randomness: %w", err)
	}

	// Format with leading zeroes (e.g. "0429")
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n.Int64()), nil
}

// HashOTP creates an uninvertible SHA-256 hex digest of the OTP.
// We never store plain OTPs in the database; even if the database is read, codes cannot be read directly.
func HashOTP(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

// VerifyOTPHash compares the input code against the stored hash in constant time.
// Go Concept - Constant Time Comparison:
// Standard string comparison `==` terminates on the first mismatched byte, creating timing variations.
// `subtle.ConstantTimeCompare` always compares all bytes regardless of mismatches, eliminating side-channel timing leaks.
func VerifyOTPHash(inputCode, storedHash string) bool {
	computed := HashOTP(inputCode)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(storedHash)) == 1
}

