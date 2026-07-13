package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// ToggleExpenseUseCase flips an expense's Active flag.
type ToggleExpenseUseCase struct {
	repo domainexpense.Repository
}

// NewToggleExpenseUseCase wires the dependencies for ToggleExpenseUseCase.
func NewToggleExpenseUseCase(repo domainexpense.Repository) *ToggleExpenseUseCase {
	return &ToggleExpenseUseCase{repo: repo}
}

// Execute loads the expense, verifies ownership, and flips Active.
func (uc *ToggleExpenseUseCase) Execute(ctx context.Context, userID, expenseID uuid.UUID) (*domainexpense.Expense, error) {
	e, err := uc.repo.FindByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar saída: %w", err)
	}
	if e.UserID != userID {
		return nil, shared.ErrNotFound
	}

	e.SetActive(!e.Active)

	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("erro ao atualizar saída: %w", err)
	}
	return e, nil
}
