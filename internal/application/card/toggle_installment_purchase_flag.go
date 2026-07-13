package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// ToggleInstallmentPurchaseFlagUseCase flips a purchase's attention flag.
type ToggleInstallmentPurchaseFlagUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

// NewToggleInstallmentPurchaseFlagUseCase wires the dependencies for
// ToggleInstallmentPurchaseFlagUseCase.
func NewToggleInstallmentPurchaseFlagUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *ToggleInstallmentPurchaseFlagUseCase {
	return &ToggleInstallmentPurchaseFlagUseCase{cards: cards, purchases: purchases}
}

// Execute loads the purchase, verifies ownership via its parent card, and
// flips its Flagged marker.
func (uc *ToggleInstallmentPurchaseFlagUseCase) Execute(ctx context.Context, userID, purchaseID uuid.UUID) (*domaincard.InstallmentPurchase, error) {
	p, err := uc.purchases.FindByID(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar compra parcelada: %w", err)
	}
	if err := verifyCardOwnership(ctx, uc.cards, p.CardID, userID); err != nil {
		return nil, err
	}

	p.SetFlagged(!p.Flagged)

	if err := uc.purchases.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao atualizar compra parcelada: %w", err)
	}
	return p, nil
}
