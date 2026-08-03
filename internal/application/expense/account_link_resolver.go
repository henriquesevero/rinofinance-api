package expense

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	domainexpense "rinofinance-api/internal/domain/expense"
)

type AccountLinkResolver struct {
	purchases domainaccount.PurchaseRepository
}

func NewAccountLinkResolver(purchases domainaccount.PurchaseRepository) *AccountLinkResolver {
	return &AccountLinkResolver{purchases: purchases}
}

func (r *AccountLinkResolver) ResolveAll(ctx context.Context, expenses []*domainexpense.Expense, reference time.Time) error {
	accountIDs := distinctAccountIDs(expenses)
	if len(accountIDs) == 0 {
		return nil
	}

	purchases, err := r.purchases.ListByAccounts(ctx, accountIDs)
	if err != nil {
		return fmt.Errorf("erro ao listar compras das contas vinculadas: %w", err)
	}

	purchasesByAccount := make(map[uuid.UUID][]*domainaccount.Purchase)
	for _, p := range purchases {
		purchasesByAccount[p.AccountID] = append(purchasesByAccount[p.AccountID], p)
	}

	for _, e := range expenses {
		if !e.IsAccountLinked() {
			continue
		}
		total := domainaccount.MonthlyPurchasesTotal(reference, purchasesByAccount[*e.AccountID])
		if err := e.SyncAmountFromAccount(total); err != nil {
			return err
		}
	}
	return nil
}

func distinctAccountIDs(expenses []*domainexpense.Expense) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	var ids []uuid.UUID
	for _, e := range expenses {
		if e.IsAccountLinked() && !seen[*e.AccountID] {
			seen[*e.AccountID] = true
			ids = append(ids, *e.AccountID)
		}
	}
	return ids
}
