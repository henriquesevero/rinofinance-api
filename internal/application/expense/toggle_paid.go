package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// TogglePaidUseCase flips an expense's Paid flag, marking whether the
// expense has actually been paid this month.
type TogglePaidUseCase struct {
	repo domainexpense.Repository
}

// NewTogglePaidUseCase wires the dependencies for TogglePaidUseCase.
func NewTogglePaidUseCase(repo domainexpense.Repository) *TogglePaidUseCase {
	return &TogglePaidUseCase{repo: repo}
}

// Execute loads the expense, verifies ownership, and flips Paid.
func (uc *TogglePaidUseCase) Execute(ctx context.Context, userID, expenseID uuid.UUID) (*domainexpense.Expense, error) {
	e, err := uc.repo.FindByID(ctx, expenseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar saída: %w", err)
	}
	if e.UserID != userID {
		return nil, shared.ErrNotFound
	}

	e.SetPaid(!e.Paid)

	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("erro ao atualizar saída: %w", err)
	}
	return e, nil
}
