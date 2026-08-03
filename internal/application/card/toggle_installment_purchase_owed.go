package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

type ToggleInstallmentPurchaseOwedUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

func NewToggleInstallmentPurchaseOwedUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *ToggleInstallmentPurchaseOwedUseCase {
	return &ToggleInstallmentPurchaseOwedUseCase{cards: cards, purchases: purchases}
}

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
