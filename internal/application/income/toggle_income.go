package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type ToggleIncomeUseCase struct {
	repo domainincome.Repository
}

func NewToggleIncomeUseCase(repo domainincome.Repository) *ToggleIncomeUseCase {
	return &ToggleIncomeUseCase{repo: repo}
}

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
