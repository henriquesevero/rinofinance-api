// Package auth holds technical (non-domain) authentication helpers:
// password hashing and JWT issuance/parsing. It is deliberately outside
// internal/domain because hashing algorithm and token format are
// infrastructure concerns, not business rules — the domain layer only
// ever sees an opaque password hash string.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plaintext password with bcrypt for storage in
// user.User.PasswordHash.
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword reports whether a plaintext password matches a stored
// bcrypt hash.
func ComparePassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// BcryptHasher adapts HashPassword/ComparePassword to the
// application/auth.PasswordHasher port, so use cases depend on an
// interface rather than this package directly.
type BcryptHasher struct{}

// Hash implements application/auth.PasswordHasher.
func (BcryptHasher) Hash(plaintext string) (string, error) {
	return HashPassword(plaintext)
}

// Compare implements application/auth.PasswordHasher.
func (BcryptHasher) Compare(hash, plaintext string) bool {
	return ComparePassword(hash, plaintext)
}
