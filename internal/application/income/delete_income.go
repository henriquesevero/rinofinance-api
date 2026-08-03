package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type DeleteIncomeUseCase struct {
	repo domainincome.Repository
}

func NewDeleteIncomeUseCase(repo domainincome.Repository) *DeleteIncomeUseCase {
	return &DeleteIncomeUseCase{repo: repo}
}

func (uc *DeleteIncomeUseCase) Execute(ctx context.Context, userID, incomeID uuid.UUID) error {
	inc, err := uc.repo.FindByID(ctx, incomeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar entrada: %w", err)
	}
	if inc.UserID != userID {
		return shared.ErrNotFound
	}

	if err := uc.repo.Delete(ctx, incomeID); err != nil {
		return fmt.Errorf("erro ao remover entrada: %w", err)
	}
	return nil
}
