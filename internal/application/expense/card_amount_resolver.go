package expense

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
)

type CardAmountResolver struct {
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

func NewCardAmountResolver(purchases domaincard.InstallmentPurchaseRepository, subscriptions domaincard.SubscriptionRepository) *CardAmountResolver {
	return &CardAmountResolver{purchases: purchases, subscriptions: subscriptions}
}

func (r *CardAmountResolver) ResolveAll(ctx context.Context, expenses []*domainexpense.Expense, reference time.Time) error {
	cardIDs := distinctCardIDs(expenses)
	if len(cardIDs) == 0 {
		return nil
	}

	purchases, err := r.purchases.ListByCards(ctx, cardIDs)
	if err != nil {
		return fmt.Errorf("erro ao listar parcelas dos cartões vinculados: %w", err)
	}
	subscriptions, err := r.subscriptions.ListByCards(ctx, cardIDs)
	if err != nil {
		return fmt.Errorf("erro ao listar assinaturas dos cartões vinculados: %w", err)
	}

	purchasesByCard := make(map[uuid.UUID][]*domaincard.InstallmentPurchase)
	for _, p := range purchases {
		purchasesByCard[p.CardID] = append(purchasesByCard[p.CardID], p)
	}
	subscriptionsByCard := make(map[uuid.UUID][]*domaincard.Subscription)
	for _, s := range subscriptions {
		subscriptionsByCard[s.CardID] = append(subscriptionsByCard[s.CardID], s)
	}

	for _, e := range expenses {
		if !e.IsCardLinked() {
			continue
		}
		total := domaincard.MonthlyTotal(reference, purchasesByCard[*e.CardID], subscriptionsByCard[*e.CardID])
		if err := e.SyncAmountFromCard(total); err != nil {
			return err
		}
	}
	return nil
}

func distinctCardIDs(expenses []*domainexpense.Expense) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	var ids []uuid.UUID
	for _, e := range expenses {
		if e.IsCardLinked() && !seen[*e.CardID] {
			seen[*e.CardID] = true
			ids = append(ids, *e.CardID)
		}
	}
	return ids
}
