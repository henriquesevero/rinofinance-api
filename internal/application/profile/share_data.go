package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	appauth "rinofinance-api/internal/application/auth"
	domainuser "rinofinance-api/internal/domain/user"
)

type ShareDataUseCase struct {
	users  domainuser.Repository
	hasher appauth.PasswordHasher
}

func NewShareDataUseCase(users domainuser.Repository, hasher appauth.PasswordHasher) *ShareDataUseCase {
	return &ShareDataUseCase{users: users, hasher: hasher}
}

func (uc *ShareDataUseCase) Execute(ctx context.Context, userID uuid.UUID, ownerEmail, ownerPassword string) (*domainuser.User, error) {
	owner, err := uc.users.FindByEmail(ctx, ownerEmail)
	if err != nil {
		return nil, domainuser.ErrInvalidCredentials
	}
	if !uc.hasher.Compare(owner.PasswordHash, ownerPassword) {
		return nil, domainuser.ErrInvalidCredentials
	}
	if owner.ID == userID {
		return nil, domainuser.ErrInvalidCredentials
	}

	me, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	me.ShareDataWith(owner.ID)
	if err := uc.users.Update(ctx, me); err != nil {
		return nil, fmt.Errorf("erro ao compartilhar conta: %w", err)
	}
	return me, nil
}

type StopSharingUseCase struct {
	users domainuser.Repository
}

func NewStopSharingUseCase(users domainuser.Repository) *StopSharingUseCase {
	return &StopSharingUseCase{users: users}
}

func (uc *StopSharingUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	me, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	me.StopSharing()
	if err := uc.users.Update(ctx, me); err != nil {
		return fmt.Errorf("erro ao desfazer compartilhamento: %w", err)
	}
	return nil
}
