package investment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

type DeleteProventoUseCase struct {
	proventos domaininvestment.ProventoRepository
}

func NewDeleteProventoUseCase(proventos domaininvestment.ProventoRepository) *DeleteProventoUseCase {
	return &DeleteProventoUseCase{proventos: proventos}
}

func (uc *DeleteProventoUseCase) Execute(ctx context.Context, userID, proventoID uuid.UUID) error {
	p, err := uc.proventos.FindByID(ctx, proventoID)
	if err != nil {
		return fmt.Errorf("erro ao buscar provento: %w", err)
	}
	if p.UserID != userID {
		return shared.ErrNotFound
	}
	if err := uc.proventos.Delete(ctx, proventoID); err != nil {
		return fmt.Errorf("erro ao remover provento: %w", err)
	}
	return nil
}
