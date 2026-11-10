// Package auth will manage user authentication, session security, and access control for Spendle.
//
// Future responsibilities:
// - User registration and login handlers
// - Password hashing and verification with bcrypt
// - JWT token generation, parsing, and claims verification
// - Authentication middleware attaching authenticated user context to requests
// - Strict cookie issuance using httpOnly, Secure, and SameSite flags
package auth

