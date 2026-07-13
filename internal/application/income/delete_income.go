package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// DeleteIncomeUseCase permanently removes an income line.
type DeleteIncomeUseCase struct {
	repo domainincome.Repository
}

// NewDeleteIncomeUseCase wires the dependencies for DeleteIncomeUseCase.
func NewDeleteIncomeUseCase(repo domainincome.Repository) *DeleteIncomeUseCase {
	return &DeleteIncomeUseCase{repo: repo}
}

// Execute verifies ownership before deleting, so a user can never delete
// another account's income by guessing its ID.
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
