package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// ToggleInstallmentPurchaseOwedUseCase flips whether a purchase counts
// toward the card's "total que devo" payoff sum.
type ToggleInstallmentPurchaseOwedUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

// NewToggleInstallmentPurchaseOwedUseCase wires the dependencies for
// ToggleInstallmentPurchaseOwedUseCase.
func NewToggleInstallmentPurchaseOwedUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *ToggleInstallmentPurchaseOwedUseCase {
	return &ToggleInstallmentPurchaseOwedUseCase{cards: cards, purchases: purchases}
}

// Execute loads the purchase, verifies ownership via its parent card, and
// flips its ExcludedFromOwed marker.
func (uc *ToggleInstallmentPurchaseOwedUseCase) Execute(ctx context.Context, userID, purchaseID uuid.UUID) (*domaincard.InstallmentPurchase, error) {
	p, err := uc.purchases.FindByID(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar compra parcelada: %w", err)
	}
	if err := verifyCardOwnership(ctx, uc.cards, p.CardID, userID); err != nil {
		return nil, err
	}

	p.SetExcludedFromOwed(!p.ExcludedFromOwed)

	if err := uc.purchases.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao atualizar compra parcelada: %w", err)
	}
	return p, nil
}
