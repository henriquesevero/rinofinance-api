package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// CreateAccountLinkedExpenseUseCase creates an expense whose amount mirrors
// an account's current-month debit purchases total (the "Compras Débito"
// rule, analogous to the card-linked expense).
type CreateAccountLinkedExpenseUseCase struct {
	expenses domainexpense.Repository
	accounts domainaccount.Repository
}

func NewCreateAccountLinkedExpenseUseCase(expenses domainexpense.Repository, accounts domainaccount.Repository) *CreateAccountLinkedExpenseUseCase {
	return &CreateAccountLinkedExpenseUseCase{expenses: expenses, accounts: accounts}
}

// Execute verifies the account belongs to the user before linking.
func (uc *CreateAccountLinkedExpenseUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID, name string, categoryID *uuid.UUID) (*domainexpense.Expense, error) {
	account, err := uc.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if account.UserID != userID {
		return nil, shared.ErrNotFound
	}

	e, err := domainexpense.NewAccountLinkedExpense(userID, name, accountID)
	if err != nil {
		return nil, err
	}
	e.SetCategory(categoryID)

	existing, err := uc.expenses.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar saídas: %w", err)
	}
	e.SetPosition(len(existing))

	if err := uc.expenses.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("erro ao criar saída vinculada à conta: %w", err)
	}
	return e, nil
}
