// Package profile orchestrates account-settings use cases: editing name
// and avatar, changing email or password, and deleting the account along
// with everything it owns.
package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainuser "rinofinance-api/internal/domain/user"
)

// UpdateProfileUseCase edits the display name and avatar. Neither
// requires the current password since they carry no account-takeover
// risk.
type UpdateProfileUseCase struct {
	users domainuser.Repository
}

// NewUpdateProfileUseCase wires the dependencies for UpdateProfileUseCase.
func NewUpdateProfileUseCase(users domainuser.Repository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{users: users}
}

// Execute renames the user and replaces their avatar (pass an empty
// avatarURL to clear it, falling back to initials in the UI).
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
