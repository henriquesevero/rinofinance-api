package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

type ReorderSubscriptionsUseCase struct {
	cards         domaincard.CardRepository
	subscriptions domaincard.SubscriptionRepository
}

func NewReorderSubscriptionsUseCase(cards domaincard.CardRepository, subscriptions domaincard.SubscriptionRepository) *ReorderSubscriptionsUseCase {
	return &ReorderSubscriptionsUseCase{cards: cards, subscriptions: subscriptions}
}

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
