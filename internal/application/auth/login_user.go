package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"rinofinance-api/internal/domain/shared"
	domainuser "rinofinance-api/internal/domain/user"
)

type LoginUserUseCase struct {
	users  domainuser.Repository
	hasher PasswordHasher
	tokens TokenIssuer
}

func NewLoginUserUseCase(users domainuser.Repository, hasher PasswordHasher, tokens TokenIssuer) *LoginUserUseCase {
	return &LoginUserUseCase{users: users, hasher: hasher, tokens: tokens}
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, email, plaintextPassword string) (string, *domainuser.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	u, err := uc.users.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return "", nil, domainuser.ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	if !uc.hasher.Compare(u.PasswordHash, plaintextPassword) {
		return "", nil, domainuser.ErrInvalidCredentials
	}

	token, err := uc.tokens.Issue(u.ID)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao emitir token: %w", err)
	}

	return token, u, nil
}
