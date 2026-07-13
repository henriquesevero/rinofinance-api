package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// UpdateIncomeUseCase renames and/or changes the amount of an existing
// income owned by the requesting user.
type UpdateIncomeUseCase struct {
	repo domainincome.Repository
}

// NewUpdateIncomeUseCase wires the dependencies for UpdateIncomeUseCase.
func NewUpdateIncomeUseCase(repo domainincome.Repository) *UpdateIncomeUseCase {
	return &UpdateIncomeUseCase{repo: repo}
}

// Execute loads the income, verifies it belongs to userID, then applies
// the new name and amount. A mismatched owner is reported as
// shared.ErrNotFound rather than a distinct "forbidden" error, so the API
// never reveals that a resource exists under another account.
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
	// Account-linked incomes have a managed amount (it mirrors the account
	// balance), so we leave it untouched and only edit name/category. The
	// amount is re-resolved on read regardless of what's stored.
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
