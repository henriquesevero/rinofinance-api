package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// ImportInstallmentInput describes one installment purchase to import from
// a parsed statement (a single "avulsa" purchase is just totalInstallments
// = 1).
type ImportInstallmentInput struct {
	Name                 string
	InstallmentAmount    shared.Money
	TotalInstallments    int
	FirstInstallmentDate time.Time
	Domain               string
	CategoryID           *uuid.UUID
}

// ImportSubscriptionInput describes one subscription to import.
type ImportSubscriptionInput struct {
	Name          string
	MonthlyAmount shared.Money
	Domain        string
	CategoryID    *uuid.UUID
}

// ImportResult reports how many items were created.
type ImportResult struct {
	InstallmentPurchases int
	Subscriptions        int
}

// ImportCardItemsUseCase bulk-creates installment purchases and
// subscriptions under a card, used by the "importar fatura" feature. It
// validates every item up front (building the domain aggregates) before
// persisting any, so a malformed entry aborts the whole import instead of
// leaving a half-imported statement.
type ImportCardItemsUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

// NewImportCardItemsUseCase wires the dependencies for
// ImportCardItemsUseCase.
func NewImportCardItemsUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
) *ImportCardItemsUseCase {
	return &ImportCardItemsUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions}
}

// Execute verifies the card belongs to userID, builds every purchase and
// subscription (failing fast on any invalid one), then persists them.
func (uc *ImportCardItemsUseCase) Execute(
	ctx context.Context,
	userID, cardID uuid.UUID,
	installments []ImportInstallmentInput,
	subscriptions []ImportSubscriptionInput,
	referenceMonth time.Time,
) (ImportResult, error) {
	if err := verifyCardOwnership(ctx, uc.cards, cardID, userID); err != nil {
		return ImportResult{}, err
	}

	// Bound imported items to the fatura's month so a mid-way parcela (e.g.
	// 2/3) shows only from the imported bill onward, never populating earlier
	// months it was never tracked in.
	effectiveFrom := referenceMonth

	// Build (and validate) everything before writing anything.
	builtPurchases := make([]*domaincard.InstallmentPurchase, 0, len(installments))
	for _, in := range installments {
		p, err := domaincard.NewInstallmentPurchase(cardID, in.Name, in.InstallmentAmount, in.TotalInstallments, in.FirstInstallmentDate)
		if err != nil {
			return ImportResult{}, fmt.Errorf("compra parcelada inválida (%q): %w", in.Name, err)
		}
		p.SetDomain(in.Domain)
		p.SetCategory(in.CategoryID)
		if !effectiveFrom.IsZero() {
			p.SetEffectiveFrom(effectiveFrom)
		}
		builtPurchases = append(builtPurchases, p)
	}

	builtSubscriptions := make([]*domaincard.Subscription, 0, len(subscriptions))
	for _, in := range subscriptions {
		s, err := domaincard.NewSubscription(cardID, in.Name, in.MonthlyAmount)
		if err != nil {
			return ImportResult{}, fmt.Errorf("assinatura inválida (%q): %w", in.Name, err)
		}
		s.SetDomain(in.Domain)
		s.SetCategory(in.CategoryID)
		if !effectiveFrom.IsZero() {
			s.SetEffectiveFrom(effectiveFrom)
		}
		builtSubscriptions = append(builtSubscriptions, s)
	}

	for _, p := range builtPurchases {
		if err := uc.purchases.Create(ctx, p); err != nil {
			return ImportResult{}, fmt.Errorf("erro ao importar compra parcelada: %w", err)
		}
	}
	for _, s := range builtSubscriptions {
		if err := uc.subscriptions.Create(ctx, s); err != nil {
			return ImportResult{}, fmt.Errorf("erro ao importar assinatura: %w", err)
		}
	}

	return ImportResult{
		InstallmentPurchases: len(builtPurchases),
		Subscriptions:        len(builtSubscriptions),
	}, nil
}
