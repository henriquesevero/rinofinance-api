package investment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// CreateProventoUseCase records a provento (dividend/rendimento) received from
// one of the user's assets.
type CreateProventoUseCase struct {
	assets    domaininvestment.Repository
	proventos domaininvestment.ProventoRepository
}

// NewCreateProventoUseCase wires the dependencies for CreateProventoUseCase.
func NewCreateProventoUseCase(assets domaininvestment.Repository, proventos domaininvestment.ProventoRepository) *CreateProventoUseCase {
	return &CreateProventoUseCase{assets: assets, proventos: proventos}
}

// Execute verifies the asset belongs to the user, then persists the provento.
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
