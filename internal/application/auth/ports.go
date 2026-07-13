// Package auth contains the use cases for account registration and login.
// It depends on the domain/user aggregate plus two small ports for the
// technical concerns of hashing and token issuance, so it can be tested
// with fakes instead of real bcrypt/JWT calls.
package auth

import "github.com/google/uuid"

// PasswordHasher hashes and verifies passwords. Implemented by
// internal/pkg/auth.BcryptHasher.
type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Compare(hash, plaintext string) bool
}

// TokenIssuer issues signed authentication tokens. Implemented by
// internal/pkg/auth.TokenIssuer.
type TokenIssuer interface {
	Issue(userID uuid.UUID) (string, error)
}
