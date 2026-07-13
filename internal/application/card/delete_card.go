package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

// DeleteCardUseCase permanently removes a credit card along with every
// installment purchase and subscription that belongs to it, and unlinks
// (rather than deletes) any expense that referenced it. This mirrors what
// a relational schema would express declaratively as ON DELETE CASCADE /
// ON DELETE SET NULL — MongoDB enforces neither automatically, so the
// application layer does it explicitly.
type DeleteCardUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
	expenses      domainexpense.Repository
}

// NewDeleteCardUseCase wires the dependencies for DeleteCardUseCase.
func NewDeleteCardUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
	expenses domainexpense.Repository,
) *DeleteCardUseCase {
	return &DeleteCardUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions, expenses: expenses}
}

// Execute verifies ownership, then deletes the card's items, unlinks its
// expenses, and finally deletes the card itself.
func (uc *DeleteCardUseCase) Execute(ctx context.Context, userID, cardID uuid.UUID) error {
	c, err := uc.cards.FindByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao buscar cartão: %w", err)
	}
	if c.UserID != userID {
		return shared.ErrNotFound
	}

	purchases, err := uc.purchases.ListByCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao listar compras parceladas do cartão: %w", err)
	}
	for _, p := range purchases {
		if err := uc.purchases.Delete(ctx, p.ID); err != nil {
			return fmt.Errorf("erro ao remover compra parcelada: %w", err)
		}
	}

	subscriptions, err := uc.subscriptions.ListByCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao listar assinaturas do cartão: %w", err)
	}
	for _, s := range subscriptions {
		if err := uc.subscriptions.Delete(ctx, s.ID); err != nil {
			return fmt.Errorf("erro ao remover assinatura: %w", err)
		}
	}

	linkedExpenses, err := uc.expenses.FindByCardID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao buscar saídas vinculadas ao cartão: %w", err)
	}
	for _, e := range linkedExpenses {
		e.Unlink()
		if err := uc.expenses.Update(ctx, e); err != nil {
			return fmt.Errorf("erro ao desvincular saída do cartão: %w", err)
		}
	}

	if err := uc.cards.Delete(ctx, cardID); err != nil {
		return fmt.Errorf("erro ao remover cartão: %w", err)
	}
	return nil
}
