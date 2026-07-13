package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// CreateSubscriptionUseCase adds a monthly subscription to a card.
type CreateSubscriptionUseCase struct {
	cards         domaincard.CardRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewCreateSubscriptionUseCase wires the dependencies for
// CreateSubscriptionUseCase.
func NewCreateSubscriptionUseCase(cards domaincard.CardRepository, subscriptions domaincard.SubscriptionRepository) *CreateSubscriptionUseCase {
	return &CreateSubscriptionUseCase{cards: cards, subscriptions: subscriptions}
}

// Execute verifies the card belongs to userID, then creates the
// subscription.
func (uc *CreateSubscriptionUseCase) Execute(ctx context.Context, userID, cardID uuid.UUID, name string, monthlyAmount shared.Money, domain string, categoryID *uuid.UUID) (*domaincard.Subscription, error) {
	if err := verifyCardOwnership(ctx, uc.cards, cardID, userID); err != nil {
		return nil, err
	}

	s, err := domaincard.NewSubscription(cardID, name, monthlyAmount)
	if err != nil {
		return nil, err
	}
	s.SetDomain(domain)
	s.SetCategory(categoryID)

	existing, err := uc.subscriptions.ListByCard(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar assinaturas do cartão: %w", err)
	}
	s.SetPosition(len(existing))

	if err := uc.subscriptions.Create(ctx, s); err != nil {
		return nil, fmt.Errorf("erro ao criar assinatura: %w", err)
	}
	return s, nil
}
