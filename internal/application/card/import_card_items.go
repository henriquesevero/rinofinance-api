package card

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type ImportInstallmentInput struct {
	Name                 string
	InstallmentAmount    shared.Money
	TotalInstallments    int
	FirstInstallmentDate time.Time
	Domain               string
	CategoryID           *uuid.UUID
}

type ImportSubscriptionInput struct {
	Name          string
	MonthlyAmount shared.Money
	Domain        string
	CategoryID    *uuid.UUID
}

type ImportResult struct {
	InstallmentPurchases int
	Subscriptions        int
}

type ImportCardItemsUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
}

func NewImportCardItemsUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
) *ImportCardItemsUseCase {
	return &ImportCardItemsUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions}
}

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

	effectiveFrom := referenceMonth

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
