package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

type DeleteCardUseCase struct {
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
	expenses      domainexpense.Repository
}

func NewDeleteCardUseCase(
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
	expenses domainexpense.Repository,
) *DeleteCardUseCase {
	return &DeleteCardUseCase{cards: cards, purchases: purchases, subscriptions: subscriptions, expenses: expenses}
}

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
