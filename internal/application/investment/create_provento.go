package investment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

type CreateProventoUseCase struct {
	assets    domaininvestment.Repository
	proventos domaininvestment.ProventoRepository
}

func NewCreateProventoUseCase(assets domaininvestment.Repository, proventos domaininvestment.ProventoRepository) *CreateProventoUseCase {
	return &CreateProventoUseCase{assets: assets, proventos: proventos}
}

func (uc *CreateProventoUseCase) Execute(ctx context.Context, userID, assetID uuid.UUID, amount shared.Money, date time.Time) (*domaininvestment.Provento, error) {
	a, err := uc.assets.FindByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ativo: %w", err)
	}
	if a.UserID != userID {
		return nil, shared.ErrNotFound
	}

	p, err := domaininvestment.NewProvento(userID, assetID, amount, date)
	if err != nil {
		return nil, err
	}
	if err := uc.proventos.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao registrar provento: %w", err)
	}
	return p, nil
}
