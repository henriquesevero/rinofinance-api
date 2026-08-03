package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

type ReorderInstallmentPurchasesUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

func NewReorderInstallmentPurchasesUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *ReorderInstallmentPurchasesUseCase {
	return &ReorderInstallmentPurchasesUseCase{cards: cards, purchases: purchases}
}

func (uc *ReorderInstallmentPurchasesUseCase) Execute(ctx context.Context, userID, cardID uuid.UUID, orderedIDs []uuid.UUID) error {
	if err := verifyCardOwnership(ctx, uc.cards, cardID, userID); err != nil {
		return err
	}

	owned, err := uc.purchases.ListByCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao listar compras do cartão: %w", err)
	}
	byID := make(map[uuid.UUID]*domaincard.InstallmentPurchase, len(owned))
	for _, p := range owned {
		byID[p.ID] = p
	}

	position := 0
	for _, id := range orderedIDs {
		p, ok := byID[id]
		if !ok {
			continue
		}
		if p.Position != position {
			p.SetPosition(position)
			if err := uc.purchases.Update(ctx, p); err != nil {
				return fmt.Errorf("erro ao reordenar compra: %w", err)
			}
		}
		position++
	}
	return nil
}
