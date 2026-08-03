package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/monthlystatus"
	"rinofinance-api/internal/domain/shared"
)

type TogglePaidUseCase struct {
	repo   domainexpense.Repository
	status monthlystatus.Repository
}

func NewTogglePaidUseCase(repo domainexpense.Repository, status monthlystatus.Repository) *TogglePaidUseCase {
	return &TogglePaidUseCase{repo: repo, status: status}
}

func (uc *TogglePaidUseCase) Execute(ctx context.Context, userID, expenseID uuid.UUID, month string) (*domainexpense.Expense, error) {
	e, err := uc.repo.FindByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar saída: %w", err)
	}
	if e.UserID != userID {
		return nil, shared.ErrNotFound
	}

	current, err := uc.status.Get(ctx, userID, monthlystatus.Expense, expenseID, month)
	if err != nil {
		return nil, err
	}
	if err := uc.status.Set(ctx, userID, monthlystatus.Expense, expenseID, month, !current); err != nil {
		return nil, err
	}

	e.SetPaid(!current)
	return e, nil
}
