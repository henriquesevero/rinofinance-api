// Package expense orchestrates CRUD, activation and card-linkage use
// cases for Aba 1's "Saídas".
package expense

import (
	"context"
	"fmt"
	"time"

	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
)

// CardAmountResolver refreshes a card-linked expense's in-memory Amount
// from its credit card's live current-month total (installment purchases
// + subscriptions). It never persists the change itself — callers decide
// whether the recomputed value should be written back, keeping this a
// pure read-time concern reusable by both the expense listing and the
// dashboard summary use cases.
type CardAmountResolver struct {
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewCardAmountResolver wires the dependencies for CardAmountResolver.
func NewCardAmountResolver(purchases domaincard.InstallmentPurchaseRepository, subscriptions domaincard.SubscriptionRepository) *CardAmountResolver {
	return &CardAmountResolver{purchases: purchases, subscriptions: subscriptions}
}

// Resolve is a no-op for standalone expenses. For card-linked expenses it
// recomputes the linked card's monthly total for reference and syncs it
// onto e in memory.
func (r *CardAmountResolver) Resolve(ctx context.Context, e *domainexpense.Expense, reference time.Time) error {
	if !e.IsCardLinked() {
		return nil
	}

	purchases, err := r.purchases.ListByCard(ctx, *e.CardID)
	if err != nil {
		return fmt.Errorf("erro ao listar parcelas do cartão vinculado: %w", err)
	}
	subscriptions, err := r.subscriptions.ListByCard(ctx, *e.CardID)
	if err != nil {
		return fmt.Errorf("erro ao listar assinaturas do cartão vinculado: %w", err)
	}

	total := domaincard.MonthlyTotal(reference, purchases, subscriptions)
	return e.SyncAmountFromCard(total)
}

// ResolveAll applies Resolve to every expense in the slice.
func (r *CardAmountResolver) ResolveAll(ctx context.Context, expenses []*domainexpense.Expense, reference time.Time) error {
	for _, e := range expenses {
		if err := r.Resolve(ctx, e, reference); err != nil {
			return err
		}
	}
	return nil
}
