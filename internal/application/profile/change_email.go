package profile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	appauth "rinofinance-api/internal/application/auth"
	"rinofinance-api/internal/domain/shared"
	domainuser "rinofinance-api/internal/domain/user"
)

type ChangeEmailUseCase struct {
	users  domainuser.Repository
	hasher appauth.PasswordHasher
}

func NewChangeEmailUseCase(users domainuser.Repository, hasher appauth.PasswordHasher) *ChangeEmailUseCase {
	return &ChangeEmailUseCase{users: users, hasher: hasher}
}

func (uc *ChangeEmailUseCase) Execute(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) (*domainuser.User, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	if !uc.hasher.Compare(u.PasswordHash, currentPassword) {
		return nil, domainuser.ErrInvalidCredentials
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(newEmail))
	existing, err := uc.users.FindByEmail(ctx, normalizedEmail)
	switch {
	case err == nil && existing.ID != userID:
		return nil, domainuser.ErrEmailAlreadyInUse
	case err != nil && !errors.Is(err, shared.ErrNotFound):
		return nil, fmt.Errorf("erro ao verificar email existente: %w", err)
	}

	if err := u.ChangeEmail(normalizedEmail); err != nil {
		return nil, err
	}
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("erro ao atualizar email: %w", err)
	}
	return u, nil
}
