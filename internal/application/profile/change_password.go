package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	appauth "rinofinance-api/internal/application/auth"
	domainuser "rinofinance-api/internal/domain/user"
)

type ChangePasswordUseCase struct {
	users  domainuser.Repository
	hasher appauth.PasswordHasher
}

func NewChangePasswordUseCase(users domainuser.Repository, hasher appauth.PasswordHasher) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{users: users, hasher: hasher}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return appauth.ErrPasswordTooShort
	}

	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	if !uc.hasher.Compare(u.PasswordHash, currentPassword) {
		return domainuser.ErrInvalidCredentials
	}

	newHash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("erro ao gerar hash de senha: %w", err)
	}
	if err := u.ChangePasswordHash(newHash); err != nil {
		return err
	}

	if err := uc.users.Update(ctx, u); err != nil {
		return fmt.Errorf("erro ao atualizar senha: %w", err)
	}
	return nil
}
