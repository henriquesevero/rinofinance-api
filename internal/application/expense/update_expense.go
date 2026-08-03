package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

type UpdateExpenseUseCase struct {
	repo domainexpense.Repository
}

func NewUpdateExpenseUseCase(repo domainexpense.Repository) *UpdateExpenseUseCase {
	return &UpdateExpenseUseCase{repo: repo}
}

func (uc *UpdateExpenseUseCase) Execute(ctx context.Context, userID, expenseID uuid.UUID, name string, amount shared.Money, categoryID *uuid.UUID) (*domainexpense.Expense, error) {
	e, err := uc.repo.FindByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar saída: %w", err)
	}
	if e.UserID != userID {
		return nil, shared.ErrNotFound
	}

	if err := e.Rename(name); err != nil {
		return nil, err
	}

	if !e.IsCardLinked() && !e.IsAccountLinked() {
		if err := e.UpdateAmount(amount); err != nil {
			return nil, err
		}
	}
	e.SetCategory(categoryID)

	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("erro ao atualizar saída: %w", err)
	}
	return e, nil
}
