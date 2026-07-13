package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// UpdateExpenseUseCase renames and/or changes the amount of an existing
// standalone expense. Card-linked and account-linked expenses reject the
// amount change (their amount is managed), surfaced as-is to the caller.
type UpdateExpenseUseCase struct {
	repo domainexpense.Repository
}

// NewUpdateExpenseUseCase wires the dependencies for UpdateExpenseUseCase.
func NewUpdateExpenseUseCase(repo domainexpense.Repository) *UpdateExpenseUseCase {
	return &UpdateExpenseUseCase{repo: repo}
}

// Execute loads the expense, verifies ownership, then applies the new name
// and amount.
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
	// Card- and account-linked expenses have a managed amount (mirrored from
	// the card total / account debits), so we leave it untouched and only
	// edit name/category. It is re-resolved on read regardless.
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
