package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type CreateAccountLinkedIncomeUseCase struct {
	incomes  domainincome.Repository
	accounts domainaccount.Repository
}

func NewCreateAccountLinkedIncomeUseCase(incomes domainincome.Repository, accounts domainaccount.Repository) *CreateAccountLinkedIncomeUseCase {
	return &CreateAccountLinkedIncomeUseCase{incomes: incomes, accounts: accounts}
}

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
