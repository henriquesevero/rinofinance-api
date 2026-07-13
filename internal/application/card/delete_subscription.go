package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// DeleteSubscriptionUseCase permanently removes a subscription.
type DeleteSubscriptionUseCase struct {
	cards         domaincard.CardRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewDeleteSubscriptionUseCase wires the dependencies for
// DeleteSubscriptionUseCase.
func NewDeleteSubscriptionUseCase(cards domaincard.CardRepository, subscriptions domaincard.SubscriptionRepository) *DeleteSubscriptionUseCase {
	return &DeleteSubscriptionUseCase{cards: cards, subscriptions: subscriptions}
}

// Execute verifies ownership (via the parent card) before deleting.
func (uc *DeleteSubscriptionUseCase) Execute(ctx context.Context, userID, subscriptionID uuid.UUID) error {
	s, err := uc.subscriptions.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar assinatura: %w", err)
	}
	if err := verifyCardOwnership(ctx, uc.cards, s.CardID, userID); err != nil {
		return err
	}

	if err := uc.subscriptions.Delete(ctx, subscriptionID); err != nil {
		return fmt.Errorf("erro ao remover assinatura: %w", err)
	}
	return nil
}
