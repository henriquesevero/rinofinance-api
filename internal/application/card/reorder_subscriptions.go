package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// ReorderSubscriptionsUseCase persists a new manual ordering of a card's
// subscriptions.
type ReorderSubscriptionsUseCase struct {
	cards         domaincard.CardRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewReorderSubscriptionsUseCase wires the dependencies.
func NewReorderSubscriptionsUseCase(cards domaincard.CardRepository, subscriptions domaincard.SubscriptionRepository) *ReorderSubscriptionsUseCase {
	return &ReorderSubscriptionsUseCase{cards: cards, subscriptions: subscriptions}
}

// Execute verifies the card belongs to the user, then assigns each of its
// subscriptions the position of its index in orderedIDs. IDs not belonging
// to the card are ignored.
func (uc *ReorderSubscriptionsUseCase) Execute(ctx context.Context, userID, cardID uuid.UUID, orderedIDs []uuid.UUID) error {
	if err := verifyCardOwnership(ctx, uc.cards, cardID, userID); err != nil {
		return err
	}

	owned, err := uc.subscriptions.ListByCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao listar assinaturas do cartão: %w", err)
	}
	byID := make(map[uuid.UUID]*domaincard.Subscription, len(owned))
	for _, s := range owned {
		byID[s.ID] = s
	}

	position := 0
	for _, id := range orderedIDs {
		s, ok := byID[id]
		if !ok {
			continue
		}
		if s.Position != position {
			s.SetPosition(position)
			if err := uc.subscriptions.Update(ctx, s); err != nil {
				return fmt.Errorf("erro ao reordenar assinatura: %w", err)
			}
		}
		position++
	}
	return nil
}
