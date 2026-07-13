package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// CreateExpenseUseCase creates a standalone expense with a manually
// entered amount (not linked to any credit card).
type CreateExpenseUseCase struct {
	repo domainexpense.Repository
}

// NewCreateExpenseUseCase wires the dependencies for CreateExpenseUseCase.
func NewCreateExpenseUseCase(repo domainexpense.Repository) *CreateExpenseUseCase {
	return &CreateExpenseUseCase{repo: repo}
}

// Execute builds and persists a new standalone Expense.
func (uc *CreateExpenseUseCase) Execute(ctx context.Context, userID uuid.UUID, name string, amount shared.Money, categoryID *uuid.UUID) (*domainexpense.Expense, error) {
	e, err := domainexpense.NewExpense(userID, name, amount)
	if err != nil {
		return nil, err
	}
	e.SetCategory(categoryID)

	existing, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar saídas: %w", err)
	}
	e.SetPosition(len(existing))

	if err := uc.repo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("erro ao criar saída: %w", err)
	}
	return e, nil
}
