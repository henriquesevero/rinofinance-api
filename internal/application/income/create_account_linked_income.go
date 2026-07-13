package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// CreateAccountLinkedIncomeUseCase creates an income whose amount mirrors a
// bank account's balance (the "vínculo com contas" rule for entradas).
type CreateAccountLinkedIncomeUseCase struct {
	incomes  domainincome.Repository
	accounts domainaccount.Repository
}

// NewCreateAccountLinkedIncomeUseCase wires the dependencies.
func NewCreateAccountLinkedIncomeUseCase(incomes domainincome.Repository, accounts domainaccount.Repository) *CreateAccountLinkedIncomeUseCase {
	return &CreateAccountLinkedIncomeUseCase{incomes: incomes, accounts: accounts}
}

// Execute verifies the account belongs to the user, then creates the
// account-linked income. The amount is left at zero; it is resolved to the
// account balance at read time.
func (uc *CreateAccountLinkedIncomeUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID, name string, categoryID *uuid.UUID) (*domainincome.Income, error) {
	account, err := uc.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if account.UserID != userID {
		return nil, shared.ErrNotFound
	}

	inc, err := domainincome.NewAccountLinkedIncome(userID, name, accountID)
	if err != nil {
		return nil, err
	}
	inc.SetCategory(categoryID)

	existing, err := uc.incomes.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar entradas: %w", err)
	}
	inc.SetPosition(len(existing))

	if err := uc.incomes.Create(ctx, inc); err != nil {
		return nil, fmt.Errorf("erro ao criar entrada vinculada à conta: %w", err)
	}
	return inc, nil
}
