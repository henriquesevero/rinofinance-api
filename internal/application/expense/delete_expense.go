package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// DeleteExpenseUseCase permanently removes an expense line.
type DeleteExpenseUseCase struct {
	repo domainexpense.Repository
}

// NewDeleteExpenseUseCase wires the dependencies for DeleteExpenseUseCase.
func NewDeleteExpenseUseCase(repo domainexpense.Repository) *DeleteExpenseUseCase {
	return &DeleteExpenseUseCase{repo: repo}
}

// Execute verifies ownership before deleting.
func (uc *DeleteExpenseUseCase) Execute(ctx context.Context, userID, expenseID uuid.UUID) error {
	e, err := uc.repo.FindByID(ctx, expenseID)
	if err != nil {
		return fmt.Errorf("erro ao buscar saída: %w", err)
	}
	if e.UserID != userID {
		return shared.ErrNotFound
	}
	if err := uc.repo.Delete(ctx, expenseID); err != nil {
		return fmt.Errorf("erro ao remover saída: %w", err)
	}
	return nil
}
