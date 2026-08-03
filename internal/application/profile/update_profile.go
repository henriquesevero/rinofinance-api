package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainuser "rinofinance-api/internal/domain/user"
)

type UpdateProfileUseCase struct {
	users domainuser.Repository
}

func NewUpdateProfileUseCase(users domainuser.Repository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{users: users}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID uuid.UUID, name, avatarURL string) (*domainuser.User, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	if err := u.Rename(name); err != nil {
		return nil, err
	}
	u.UpdateAvatar(avatarURL)

	if err := uc.users.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("erro ao atualizar perfil: %w", err)
	}
	return u, nil
}
