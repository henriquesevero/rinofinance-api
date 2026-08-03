package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
)

type CreateAssetUseCase struct {
	repo domaininvestment.Repository
}

func NewCreateAssetUseCase(repo domaininvestment.Repository) *CreateAssetUseCase {
	return &CreateAssetUseCase{repo: repo}
}

func (uc *CreateAssetUseCase) Execute(ctx context.Context, userID uuid.UUID, in domaininvestment.AssetInput) (*domaininvestment.Asset, error) {
	a, err := domaininvestment.NewAsset(userID, in)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao criar ativo: %w", err)
	}
	return a, nil
}
