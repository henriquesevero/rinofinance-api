package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type UpdateInstallmentPurchaseUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

func NewUpdateInstallmentPurchaseUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *UpdateInstallmentPurchaseUseCase {
	return &UpdateInstallmentPurchaseUseCase{cards: cards, purchases: purchases}
}

func (uc *UpdateInstallmentPurchaseUseCase) Execute(
	ctx context.Context,
	userID, purchaseID uuid.UUID,
	name string,
	installmentAmount shared.Money,
	totalInstallments int,
	firstInstallmentDate time.Time,
	domain string,
	categoryID *uuid.UUID,
) (*domaincard.InstallmentPurchase, error) {
	p, err := uc.purchases.FindByID(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar compra parcelada: %w", err)
	}
	if err := verifyCardOwnership(ctx, uc.cards, p.CardID, userID); err != nil {
		return nil, err
	}

	rebuilt, err := domaincard.NewInstallmentPurchase(p.CardID, name, installmentAmount, totalInstallments, firstInstallmentDate)
	if err != nil {
		return nil, err
	}
	rebuilt.ID = p.ID
	rebuilt.CreatedAt = p.CreatedAt
	rebuilt.SetDomain(domain)
	rebuilt.SetCategory(categoryID)

	rebuilt.SetFlagged(p.Flagged)

	if err := uc.purchases.Update(ctx, rebuilt); err != nil {
		return nil, fmt.Errorf("erro ao atualizar compra parcelada: %w", err)
	}
	return rebuilt, nil
}
