package expense

import (
	"context"
	"fmt"
	"time"

	domainaccount "rinofinance-api/internal/domain/account"
	domainexpense "rinofinance-api/internal/domain/expense"
)

// AccountLinkResolver refreshes an account-linked expense's in-memory
// Amount from its account's current-month debit purchases total. Like
// CardAmountResolver, it never persists — a read-time concern used by the
// expense listing and the dashboard summary.
type AccountLinkResolver struct {
	purchases domainaccount.PurchaseRepository
}

// NewAccountLinkResolver wires the dependencies.
func NewAccountLinkResolver(purchases domainaccount.PurchaseRepository) *AccountLinkResolver {
	return &AccountLinkResolver{purchases: purchases}
}

// Resolve is a no-op for non-account-linked expenses. For account-linked
// ones it recomputes the account's monthly debit total for reference.
func (r *AccountLinkResolver) Resolve(ctx context.Context, e *domainexpense.Expense, reference time.Time) error {
	if !e.IsAccountLinked() {
		return nil
	}
	purchases, err := r.purchases.ListByAccount(ctx, *e.AccountID)
	if err != nil {
		return fmt.Errorf("erro ao listar compras da conta vinculada: %w", err)
	}
	total := domainaccount.MonthlyPurchasesTotal(reference, purchases)
	return e.SyncAmountFromAccount(total)
}

// ResolveAll applies Resolve to every expense in the slice.
func (r *AccountLinkResolver) ResolveAll(ctx context.Context, expenses []*domainexpense.Expense, reference time.Time) error {
	for _, e := range expenses {
		if err := r.Resolve(ctx, e, reference); err != nil {
			return err
		}
	}
	return nil
}
