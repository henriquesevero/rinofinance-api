package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// ToggleIncomeUseCase flips an income's Active flag, implementing the
// "Regra de Ativação" — the value stops counting toward the monthly sum
// without being deleted.
type ToggleIncomeUseCase struct {
	repo domainincome.Repository
}

// NewToggleIncomeUseCase wires the dependencies for ToggleIncomeUseCase.
func NewToggleIncomeUseCase(repo domainincome.Repository) *ToggleIncomeUseCase {
	return &ToggleIncomeUseCase{repo: repo}
}

// Execute loads the income, verifies ownership, and flips Active.
func (uc *ToggleIncomeUseCase) Execute(ctx context.Context, userID, incomeID uuid.UUID) (*domainincome.Income, error) {
	inc, err := uc.repo.FindByID(ctx, incomeID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar entrada: %w", err)
	}
	if inc.UserID != userID {
		return nil, shared.ErrNotFound
	}

	inc.SetActive(!inc.Active)

	if err := uc.repo.Update(ctx, inc); err != nil {
		return nil, fmt.Errorf("erro ao atualizar entrada: %w", err)
	}
	return inc, nil
}
