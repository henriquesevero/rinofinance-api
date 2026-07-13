package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// UpdateInstallmentPurchaseUseCase edits an existing installment purchase.
type UpdateInstallmentPurchaseUseCase struct {
	cards     domaincard.CardRepository
	purchases domaincard.InstallmentPurchaseRepository
}

// NewUpdateInstallmentPurchaseUseCase wires the dependencies for
// UpdateInstallmentPurchaseUseCase.
func NewUpdateInstallmentPurchaseUseCase(cards domaincard.CardRepository, purchases domaincard.InstallmentPurchaseRepository) *UpdateInstallmentPurchaseUseCase {
	return &UpdateInstallmentPurchaseUseCase{cards: cards, purchases: purchases}
}

// Execute loads the purchase, verifies (via its parent card) that it
// belongs to userID, then replaces its fields.
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
	// Editing an installment purchase must not clear an attention flag the
	// user set earlier, so carry it over from the stored purchase.
	rebuilt.SetFlagged(p.Flagged)

	if err := uc.purchases.Update(ctx, rebuilt); err != nil {
		return nil, fmt.Errorf("erro ao atualizar compra parcelada: %w", err)
	}
	return rebuilt, nil
}
