package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

type DeleteInstallmentPurchaseUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

func NewDeleteInstallmentPurchaseUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *DeleteInstallmentPurchaseUseCase {
	return &DeleteInstallmentPurchaseUseCase{cards: cards, purchases: purchases}
}

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
