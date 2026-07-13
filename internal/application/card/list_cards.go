package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// CardOverview bundles a credit card with its items and current-month
// total, the shape Aba 2 needs to render one card's section.
type CardOverview struct {
	Card                 *domaincard.CreditCard
	InstallmentPurchases []*domaincard.InstallmentPurchase
	Subscriptions        []*domaincard.Subscription
	MonthlyTotal         shared.Money
}

// ListCardsUseCase lists every credit card belonging to a user along with
// each card's current-month total and the combined "Total Geral" across
// all cards.
type ListCardsUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewListCardsUseCase wires the dependencies for ListCardsUseCase.
func NewListCardsUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
) *ListCardsUseCase {
	return &ListCardsUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions}
}

// Execute returns one CardOverview per card plus the grand total across
// all of them, both computed as of the reference month.
func (uc *ListCardsUseCase) Execute(ctx context.Context, userID uuid.UUID, reference time.Time) ([]CardOverview, shared.Money, error) {
	cards, err := uc.cards.ListByUser(ctx, userID)
	if err != nil {
		return nil, shared.Zero, fmt.Errorf("erro ao listar cartões: %w", err)
	}

	overviews := make([]CardOverview, 0, len(cards))
	grandTotal := shared.Zero

	for _, c := range cards {
		purchases, err := uc.purchases.ListByCard(ctx, c.ID)
		if err != nil {
			return nil, shared.Zero, fmt.Errorf("erro ao listar parcelas do cartão %s: %w", c.Name, err)
		}
		subscriptions, err := uc.subscriptions.ListByCard(ctx, c.ID)
		if err != nil {
			return nil, shared.Zero, fmt.Errorf("erro ao listar assinaturas do cartão %s: %w", c.Name, err)
		}

		total := domaincard.MonthlyTotal(reference, purchases, subscriptions)
		grandTotal = grandTotal.Add(total)

		overviews = append(overviews, CardOverview{
			Card:                 c,
			InstallmentPurchases: purchases,
			Subscriptions:        subscriptions,
			MonthlyTotal:         total,
		})
	}

	return overviews, grandTotal, nil
}
