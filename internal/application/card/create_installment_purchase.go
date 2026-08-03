package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type CreateInstallmentPurchaseUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

func NewCreateInstallmentPurchaseUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *CreateInstallmentPurchaseUseCase {
	return &CreateInstallmentPurchaseUseCase{cards: cards, purchases: purchases}
}

func (uc *CreateInstallmentPurchaseUseCase) Execute(
	ctx context.Context,
	userID, cardID uuid.UUID,
	name string,
	installmentAmount shared.Money,
	totalInstallments int,
	firstInstallmentDate time.Time,
	domain string,
	categoryID *uuid.UUID,
) (*domaincard.InstallmentPurchase, error) {
	if err := verifyCardOwnership(ctx, uc.cards, cardID, userID); err != nil {
		return nil, err
	}

	p, err := domaincard.NewInstallmentPurchase(cardID, name, installmentAmount, totalInstallments, firstInstallmentDate)
	if err != nil {
		return nil, err
	}
	p.SetDomain(domain)
	p.SetCategory(categoryID)

	existing, err := uc.purchases.ListByCard(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar compras do cartão: %w", err)
	}
	p.SetPosition(len(existing))

	if err := uc.purchases.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao criar compra parcelada: %w", err)
	}
	return p, nil
}
