package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type UpdateSubscriptionUseCase struct {
	cards         domaincard.CardRepository
	subscriptions domaincard.SubscriptionRepository
}

func NewUpdateSubscriptionUseCase(cards domaincard.CardRepository, subscriptions domaincard.SubscriptionRepository) *UpdateSubscriptionUseCase {
	return &UpdateSubscriptionUseCase{cards: cards, subscriptions: subscriptions}
}

func (uc *UpdateSubscriptionUseCase) Execute(ctx context.Context, userID, subscriptionID uuid.UUID, name string, monthlyAmount shared.Money, domain string, categoryID *uuid.UUID) (*domaincard.Subscription, error) {
	s, err := uc.subscriptions.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar assinatura: %w", err)
	}
	if err := verifyCardOwnership(ctx, uc.cards, s.CardID, userID); err != nil {
		return nil, err
	}

	if err := s.Rename(name); err != nil {
		return nil, err
	}
	if err := s.UpdateMonthlyAmount(monthlyAmount); err != nil {
		return nil, err
	}
	s.SetDomain(domain)
	s.SetCategory(categoryID)

	if err := uc.subscriptions.Update(ctx, s); err != nil {
		return nil, fmt.Errorf("erro ao atualizar assinatura: %w", err)
	}
	return s, nil
}
