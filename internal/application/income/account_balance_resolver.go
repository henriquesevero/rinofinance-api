package income

import (
	"context"
	"errors"
	"fmt"

	domainaccount "rinofinance-api/internal/domain/account"
	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type AccountBalanceResolver struct {
	accounts domainaccount.Repository
}

func NewAccountBalanceResolver(accounts domainaccount.Repository) *AccountBalanceResolver {
	return &AccountBalanceResolver{accounts: accounts}
}

func (r *AccountBalanceResolver) Resolve(ctx context.Context, inc *domainincome.Income) error {
	if !inc.IsAccountLinked() {
		return nil
	}
	account, err := r.accounts.FindByID(ctx, *inc.AccountID)
	if err != nil {

		if errors.Is(err, shared.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("erro ao buscar conta vinculada: %w", err)
	}
	return inc.SyncAmountFromAccount(account.Balance)
}

func (r *AccountBalanceResolver) ResolveAll(ctx context.Context, incomes []*domainincome.Income) error {
	for _, inc := range incomes {
		if err := r.Resolve(ctx, inc); err != nil {
			return err
		}
	}
	return nil
}
