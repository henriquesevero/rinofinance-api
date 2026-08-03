package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

type UpdateAssetUseCase struct {
	repo domaininvestment.Repository
}

func NewUpdateAssetUseCase(repo domaininvestment.Repository) *UpdateAssetUseCase {
	return &UpdateAssetUseCase{repo: repo}
}

func (uc *UpdateAssetUseCase) Execute(ctx context.Context, userID, assetID uuid.UUID, in domaininvestment.AssetInput) (*domaininvestment.Asset, error) {
	a, err := uc.repo.FindByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	if a.UserID != userID {
		return nil, shared.ErrNotFound
	}

	if err := a.Update(in); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao atualizar ativo: %w", err)
	}
	return a, nil
}
