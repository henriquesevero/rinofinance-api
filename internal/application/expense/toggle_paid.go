package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/monthlystatus"
	"rinofinance-api/internal/domain/shared"
)

// TogglePaidUseCase flips whether an expense was paid in a given month. The
// flag is stored per month (not on the expense itself), so it resets every
// month and reflects the month being viewed.
type TogglePaidUseCase struct {
	repo   domainexpense.Repository
	status monthlystatus.Repository
}

// NewTogglePaidUseCase wires the dependencies for TogglePaidUseCase.
func NewTogglePaidUseCase(repo domainexpense.Repository, status monthlystatus.Repository) *TogglePaidUseCase {
	return &TogglePaidUseCase{repo: repo, status: status}
}

// Execute verifies ownership, flips the month's paid status, and returns the
// expense carrying that month's value.
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

	e.SetPaid(!current) // reflect the month's status on the returned entity
	return e, nil
}
