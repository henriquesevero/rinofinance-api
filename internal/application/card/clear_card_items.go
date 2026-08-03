package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

type ClearMode string

const (
	ClearModeEnd ClearMode = "end"

	ClearModeDelete ClearMode = "delete"
)

type ClearResult struct {
	InstallmentPurchases int
	Subscriptions        int
}

type ClearCardItemsUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

func NewClearCardItemsUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
) *ClearCardItemsUseCase {
	return &ClearCardItemsUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions}
}

func (uc *ClearCardItemsUseCase) Execute(
	ctx context.Context,
	userID, cardID uuid.UUID,
	purchaseIDs, subscriptionIDs []uuid.UUID,
	mode ClearMode,
	month time.Time,
) (ClearResult, error) {
	if err := verifyCardOwnership(ctx, uc.cards, cardID, userID); err != nil {
		return ClearResult{}, err
	}

	result := ClearResult{}

	for _, id := range purchaseIDs {
		p, err := uc.purchases.FindByID(ctx, id)
		if err != nil {
			continue
		}
		if p.CardID != cardID {
			continue
		}
		if mode == ClearModeEnd {
			p.EndFrom(month)
			if err := uc.purchases.Update(ctx, p); err != nil {
				return result, fmt.Errorf("erro ao encerrar compra parcelada: %w", err)
			}
		} else if err := uc.purchases.Delete(ctx, id); err != nil {
			return result, fmt.Errorf("erro ao remover compra parcelada: %w", err)
		}
		result.InstallmentPurchases++
	}

	for _, id := range subscriptionIDs {
		s, err := uc.subscriptions.FindByID(ctx, id)
		if err != nil {
			continue
		}
		if s.CardID != cardID {
			continue
		}
		if mode == ClearModeEnd {
			s.EndFrom(month)
			if err := uc.subscriptions.Update(ctx, s); err != nil {
				return result, fmt.Errorf("erro ao encerrar assinatura: %w", err)
			}
		} else if err := uc.subscriptions.Delete(ctx, id); err != nil {
			return result, fmt.Errorf("erro ao remover assinatura: %w", err)
		}
		result.Subscriptions++
	}

	return result, nil
}
