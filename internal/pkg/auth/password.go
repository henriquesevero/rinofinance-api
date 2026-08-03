package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

type BcryptHasher struct{}

func (BcryptHasher) Hash(plaintext string) (string, error) {
	return HashPassword(plaintext)
}

func (BcryptHasher) Compare(hash, plaintext string) bool {
	return ComparePassword(hash, plaintext)
}
