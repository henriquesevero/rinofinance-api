package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

type PortfolioOverview struct {
	Assets         []*domaininvestment.Asset
	Proventos      []*domaininvestment.Provento
	TotalPatrimony shared.Money
	TotalInvested  shared.Money
	TotalProventos shared.Money
}

type ListAssetsUseCase struct {
	assets    domaininvestment.Repository
	proventos domaininvestment.ProventoRepository
}

func NewListAssetsUseCase(assets domaininvestment.Repository, proventos domaininvestment.ProventoRepository) *ListAssetsUseCase {
	return &ListAssetsUseCase{assets: assets, proventos: proventos}
}

func (uc *ListAssetsUseCase) Execute(ctx context.Context, userID uuid.UUID) (PortfolioOverview, error) {
	assets, err := uc.assets.ListByUser(ctx, userID)
	if err != nil {
		return PortfolioOverview{}, fmt.Errorf("erro ao listar ativos: %w", err)
	}
	proventos, err := uc.proventos.ListByUser(ctx, userID)
	if err != nil {
		return PortfolioOverview{}, fmt.Errorf("erro ao listar proventos: %w", err)
	}
	return PortfolioOverview{
		Assets:         assets,
		Proventos:      proventos,
		TotalPatrimony: domaininvestment.TotalPatrimony(assets),
		TotalInvested:  domaininvestment.TotalInvested(assets),
		TotalProventos: domaininvestment.TotalProventos(proventos),
	}, nil
}
