package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	appauth "rinofinance-api/internal/application/auth"
	domainuser "rinofinance-api/internal/domain/user"
)

// ShareDataUseCase links the current account to another account so both see
// and edit the same data (shared household). The current user proves access
// to the target account with its email + password.
type ShareDataUseCase struct {
	users  domainuser.Repository
	hasher appauth.PasswordHasher
}

// NewShareDataUseCase wires the dependencies for ShareDataUseCase.
func NewShareDataUseCase(users domainuser.Repository, hasher appauth.PasswordHasher) *ShareDataUseCase {
	return &ShareDataUseCase{users: users, hasher: hasher}
}

// Execute points userID's account at ownerEmail's data after verifying the
// owner's password.
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

// StopSharingUseCase reverts the account to its own data.
type StopSharingUseCase struct {
	users domainuser.Repository
}

// NewStopSharingUseCase wires the dependencies for StopSharingUseCase.
func NewStopSharingUseCase(users domainuser.Repository) *StopSharingUseCase {
	return &StopSharingUseCase{users: users}
}

// Execute clears the current user's data-owner link.
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
