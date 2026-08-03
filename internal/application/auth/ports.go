package auth

import "github.com/google/uuid"

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Compare(hash, plaintext string) bool
}

type TokenIssuer interface {
	Issue(userID uuid.UUID) (string, error)
}
