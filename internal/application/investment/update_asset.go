package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// UpdateAssetUseCase renames and/or updates the balance of an existing
// asset.
type UpdateAssetUseCase struct {
	repo domaininvestment.Repository
}

// NewUpdateAssetUseCase wires the dependencies for UpdateAssetUseCase.
func NewUpdateAssetUseCase(repo domaininvestment.Repository) *UpdateAssetUseCase {
	return &UpdateAssetUseCase{repo: repo}
}

// Execute loads the asset, verifies ownership, then applies the new name
// and balance.
func (uc *UpdateAssetUseCase) Execute(ctx context.Context, userID, assetID uuid.UUID, name string, currentBalance shared.Money) (*domaininvestment.Asset, error) {
	a, err := uc.repo.FindByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	if a.UserID != userID {
		return nil, shared.ErrNotFound
	}

	if err := a.Rename(name); err != nil {
		return nil, err
	}
	if err := a.UpdateBalance(currentBalance); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao atualizar ativo: %w", err)
	}
	return a, nil
}
