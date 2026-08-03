package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type UpdateIncomeUseCase struct {
	repo domainincome.Repository
}

func NewUpdateIncomeUseCase(repo domainincome.Repository) *UpdateIncomeUseCase {
	return &UpdateIncomeUseCase{repo: repo}
}

func (uc *UpdateIncomeUseCase) Execute(ctx context.Context, userID, incomeID uuid.UUID, name string, amount shared.Money, categoryID *uuid.UUID) (*domainincome.Income, error) {
	inc, err := uc.repo.FindByID(ctx, incomeID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar entrada: %w", err)
	}
	if inc.UserID != userID {
		return nil, shared.ErrNotFound
	}

	if err := inc.Rename(name); err != nil {
		return nil, err
	}

	if !inc.IsAccountLinked() {
		if err := inc.UpdateAmount(amount); err != nil {
			return nil, err
		}
	}
	inc.SetCategory(categoryID)

	if err := uc.repo.Update(ctx, inc); err != nil {
		return nil, fmt.Errorf("erro ao atualizar entrada: %w", err)
	}
	return inc, nil
}
