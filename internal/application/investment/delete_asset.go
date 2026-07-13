package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// DeleteAssetUseCase permanently removes an investment/patrimony asset.
type DeleteAssetUseCase struct {
	repo domaininvestment.Repository
}

// NewDeleteAssetUseCase wires the dependencies for DeleteAssetUseCase.
func NewDeleteAssetUseCase(repo domaininvestment.Repository) *DeleteAssetUseCase {
	return &DeleteAssetUseCase{repo: repo}
}

// Execute verifies ownership before deleting.
func (uc *DeleteAssetUseCase) Execute(ctx context.Context, userID, assetID uuid.UUID) error {
	a, err := uc.repo.FindByID(ctx, assetID)
	if err != nil {
		return fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	if a.UserID != userID {
		return shared.ErrNotFound
	}

	if err := uc.repo.Delete(ctx, assetID); err != nil {
		return fmt.Errorf("erro ao remover ativo: %w", err)
	}
	return nil
}
