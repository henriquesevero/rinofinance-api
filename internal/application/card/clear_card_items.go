package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// ClearMode selects how "limpar fatura" removes items.
type ClearMode string

const (
	// ClearModeEnd stops the item from the reference month onward while
	// keeping every earlier month's bill intact (preserves history).
	ClearModeEnd ClearMode = "end"
	// ClearModeDelete permanently deletes the item from every month.
	ClearModeDelete ClearMode = "delete"
)

// ClearResult reports how many items were removed.
type ClearResult struct {
	InstallmentPurchases int
	Subscriptions        int
}

// ClearCardItemsUseCase bulk-deletes selected installment purchases and
// subscriptions from a card, used by the "limpar fatura" feature. Each
// item is verified to actually belong to the target card before deletion,
// so a crafted ID from another card can't be removed.
type ClearCardItemsUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewClearCardItemsUseCase wires the dependencies for
// ClearCardItemsUseCase.
func NewClearCardItemsUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
) *ClearCardItemsUseCase {
	return &ClearCardItemsUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions}
}

// Execute verifies the card belongs to userID, then removes each listed
// purchase/subscription that actually belongs to that card. With
// ClearModeEnd it ends items from the reference month onward (keeping the
// past); with ClearModeDelete it deletes them outright. IDs that don't
// belong to the card (or no longer exist) are silently ignored, keeping the
// operation idempotent.
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
			continue // already gone
		}
		if p.CardID != cardID {
			continue // belongs to another card
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
