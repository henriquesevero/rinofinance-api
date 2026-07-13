package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// ListAssetsUseCase lists every investment/patrimony asset belonging to a
// user, along with the total patrimony card shown in Aba 3.
type ListAssetsUseCase struct {
	repo domaininvestment.Repository
}

// NewListAssetsUseCase wires the dependencies for ListAssetsUseCase.
func NewListAssetsUseCase(repo domaininvestment.Repository) *ListAssetsUseCase {
	return &ListAssetsUseCase{repo: repo}
}

// Execute returns all assets for userID plus their combined total balance.
func (uc *ListAssetsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*domaininvestment.Asset, shared.Money, error) {
	assets, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, shared.Zero, fmt.Errorf("erro ao listar ativos: %w", err)
	}
	return assets, domaininvestment.TotalPatrimony(assets), nil
}
