package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"rinofinance-api/internal/domain/shared"
	domainuser "rinofinance-api/internal/domain/user"
)

// RegisterUserUseCase creates a new account with a hashed password.
type RegisterUserUseCase struct {
	users  domainuser.Repository
	hasher PasswordHasher
}

// NewRegisterUserUseCase wires the dependencies for RegisterUserUseCase.
func NewRegisterUserUseCase(users domainuser.Repository, hasher PasswordHasher) *RegisterUserUseCase {
	return &RegisterUserUseCase{users: users, hasher: hasher}
}

// Execute validates the password strength, ensures the email isn't already
// registered, hashes the password and persists the new user.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, name, email, plaintextPassword string) (*domainuser.User, error) {
	if len(plaintextPassword) < 8 {
		return nil, ErrPasswordTooShort
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	_, err := uc.users.FindByEmail(ctx, normalizedEmail)
	switch {
	case err == nil:
		return nil, domainuser.ErrEmailAlreadyInUse
	case !errors.Is(err, shared.ErrNotFound):
		return nil, fmt.Errorf("erro ao verificar email existente: %w", err)
	}

	hash, err := uc.hasher.Hash(plaintextPassword)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash de senha: %w", err)
	}

	u, err := domainuser.NewUser(name, normalizedEmail, hash)
	if err != nil {
		return nil, err
	}

	if err := uc.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("erro ao criar usuário: %w", err)
	}

	return u, nil
}
