package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type CardOverview struct {
	Card                 *domaincard.CreditCard
	InstallmentPurchases []*domaincard.InstallmentPurchase
	Subscriptions        []*domaincard.Subscription
	MonthlyTotal         shared.Money
}

type ListCardsUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

func NewListCardsUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
) *ListCardsUseCase {
	return &ListCardsUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions}
}

func (uc *ListCardsUseCase) Execute(ctx context.Context, userID uuid.UUID, reference time.Time) ([]CardOverview, shared.Money, error) {
	cards, err := uc.cards.ListByUser(ctx, userID)
	if err != nil {
		return nil, shared.Zero, fmt.Errorf("erro ao listar cartões: %w", err)
	}

	cardIDs := make([]uuid.UUID, len(cards))
	for i, c := range cards {
		cardIDs[i] = c.ID
	}

	allPurchases, err := uc.purchases.ListByCards(ctx, cardIDs)
	if err != nil {
		return nil, shared.Zero, fmt.Errorf("erro ao listar parcelas: %w", err)
	}
	allSubscriptions, err := uc.subscriptions.ListByCards(ctx, cardIDs)
	if err != nil {
		return nil, shared.Zero, fmt.Errorf("erro ao listar assinaturas: %w", err)
	}

	purchasesByCard := groupByCard(allPurchases, func(p *domaincard.InstallmentPurchase) uuid.UUID { return p.CardID })
	subscriptionsByCard := groupByCard(allSubscriptions, func(s *domaincard.Subscription) uuid.UUID { return s.CardID })

	overviews := make([]CardOverview, 0, len(cards))
	grandTotal := shared.Zero
	for _, c := range cards {
		purchases := purchasesByCard[c.ID]
		subscriptions := subscriptionsByCard[c.ID]
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

func groupByCard[T any](items []T, cardID func(T) uuid.UUID) map[uuid.UUID][]T {
	grouped := make(map[uuid.UUID][]T)
	for _, item := range items {
		key := cardID(item)
		grouped[key] = append(grouped[key], item)
	}
	return grouped
}
