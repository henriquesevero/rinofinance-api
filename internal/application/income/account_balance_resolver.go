package income

import (
	"context"
	"errors"
	"fmt"

	domainaccount "rinofinance-api/internal/domain/account"
	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// AccountBalanceResolver refreshes an account-linked income's in-memory
// Amount from its bank account's live balance. It never persists the
// change — a pure read-time concern reused by the income listing and the
// dashboard summary, mirroring the expense CardAmountResolver.
type AccountBalanceResolver struct {
	accounts domainaccount.Repository
}

// NewAccountBalanceResolver wires the dependencies.
func NewAccountBalanceResolver(accounts domainaccount.Repository) *AccountBalanceResolver {
	return &AccountBalanceResolver{accounts: accounts}
}

// Resolve is a no-op for standalone incomes. For account-linked incomes it
// syncs the linked account's current balance onto inc in memory.
func (r *AccountBalanceResolver) Resolve(ctx context.Context, inc *domainincome.Income) error {
	if !inc.IsAccountLinked() {
		return nil
	}
	account, err := r.accounts.FindByID(ctx, *inc.AccountID)
	if err != nil {
		// A deleted account shouldn't break the income list; leave the last
		// synced amount in place.
		if errors.Is(err, shared.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("erro ao buscar conta vinculada: %w", err)
	}
	return inc.SyncAmountFromAccount(account.Balance)
}

// ResolveAll applies Resolve to every income in the slice.
func (r *AccountBalanceResolver) ResolveAll(ctx context.Context, incomes []*domainincome.Income) error {
	for _, inc := range incomes {
		if err := r.Resolve(ctx, inc); err != nil {
			return err
		}
	}
	return nil
}
