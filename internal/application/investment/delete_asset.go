package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// DeleteAssetUseCase permanently removes an asset and its proventos.
type DeleteAssetUseCase struct {
	assets    domaininvestment.Repository
	proventos domaininvestment.ProventoRepository
}

// NewDeleteAssetUseCase wires the dependencies for DeleteAssetUseCase.
func NewDeleteAssetUseCase(assets domaininvestment.Repository, proventos domaininvestment.ProventoRepository) *DeleteAssetUseCase {
	return &DeleteAssetUseCase{assets: assets, proventos: proventos}
}

// Execute verifies ownership before deleting the asset and every provento
// tied to it.
func (uc *DeleteAssetUseCase) Execute(ctx context.Context, userID, assetID uuid.UUID) error {
	a, err := uc.assets.FindByID(ctx, assetID)
	if err != nil {
		return fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	if a.UserID != userID {
		return shared.ErrNotFound
	}

	if err := uc.proventos.DeleteByAsset(ctx, assetID); err != nil {
		return fmt.Errorf("erro ao remover proventos do ativo: %w", err)
	}
	if err := uc.assets.Delete(ctx, assetID); err != nil {
		return fmt.Errorf("erro ao remover ativo: %w", err)
	}
	return nil
}
