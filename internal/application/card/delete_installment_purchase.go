package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// DeleteInstallmentPurchaseUseCase permanently removes an installment
// purchase.
type DeleteInstallmentPurchaseUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

// NewDeleteInstallmentPurchaseUseCase wires the dependencies for
// DeleteInstallmentPurchaseUseCase.
func NewDeleteInstallmentPurchaseUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *DeleteInstallmentPurchaseUseCase {
	return &DeleteInstallmentPurchaseUseCase{cards: cards, purchases: purchases}
}

// Execute verifies ownership (via the parent card) before deleting.
func (uc *DeleteInstallmentPurchaseUseCase) Execute(ctx context.Context, userID, purchaseID uuid.UUID) error {
	p, err := uc.purchases.FindByID(ctx, purchaseID)
	if err != nil {
		return fmt.Errorf("erro ao buscar compra parcelada: %w", err)
	}
	if err := verifyCardOwnership(ctx, uc.cards, p.CardID, userID); err != nil {
		return err
	}

	if err := uc.purchases.Delete(ctx, purchaseID); err != nil {
		return fmt.Errorf("erro ao remover compra parcelada: %w", err)
	}
	return nil
}
