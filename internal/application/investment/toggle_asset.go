package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// ToggleAssetUseCase flips an asset's Active flag, so its balance stops
// counting toward the total patrimony without being deleted.
type ToggleAssetUseCase struct {
	repo domaininvestment.Repository
}

// NewToggleAssetUseCase wires the dependencies for ToggleAssetUseCase.
func NewToggleAssetUseCase(repo domaininvestment.Repository) *ToggleAssetUseCase {
	return &ToggleAssetUseCase{repo: repo}
}

// Execute loads the asset, verifies ownership, and flips Active.
func (uc *ToggleAssetUseCase) Execute(ctx context.Context, userID, assetID uuid.UUID) (*domaininvestment.Asset, error) {
	a, err := uc.repo.FindByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	if a.UserID != userID {
		return nil, shared.ErrNotFound
	}

	a.SetActive(!a.Active)

	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao atualizar ativo: %w", err)
	}
	return a, nil
}
