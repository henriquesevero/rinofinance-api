package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	appauth "rinofinance-api/internal/application/auth"
	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	domainincome "rinofinance-api/internal/domain/income"
	domaininvestment "rinofinance-api/internal/domain/investment"
	domainuser "rinofinance-api/internal/domain/user"
)

// DeleteAccountUseCase permanently deletes a user and every aggregate
// they own — incomes, expenses, credit cards (with their installment
// purchases and subscriptions), and investment assets. It requires the
// current password so a hijacked session token alone can't wipe an
// account. MongoDB has no cascading foreign keys, so this orchestrates
// the deletion explicitly, mirroring application/card.DeleteCardUseCase's
// approach for a single card.
type DeleteAccountUseCase struct {
	users         domainuser.Repository
	hasher        appauth.PasswordHasher
	incomes       domainincome.Repository
	expenses      domainexpense.Repository
	cards         domaincard.CardRepository
	purchases     domaincard.InstallmentPurchaseRepository
	subscriptions domaincard.SubscriptionRepository
	investments   domaininvestment.Repository
}

// NewDeleteAccountUseCase wires the dependencies for
// DeleteAccountUseCase.
func NewDeleteAccountUseCase(
	users domainuser.Repository,
	hasher appauth.PasswordHasher,
	incomes domainincome.Repository,
	expenses domainexpense.Repository,
	cards domaincard.CardRepository,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
	investments domaininvestment.Repository,
) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{
		users:         users,
		hasher:        hasher,
		incomes:       incomes,
		expenses:      expenses,
		cards:         cards,
		purchases:     purchases,
		subscriptions: subscriptions,
		investments:   investments,
	}
}

// Execute verifies currentPassword, deletes every aggregate owned by
// userID, then deletes the user itself.
func (uc *DeleteAccountUseCase) Execute(ctx context.Context, userID uuid.UUID, currentPassword string) error {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	if !uc.hasher.Compare(u.PasswordHash, currentPassword) {
		return domainuser.ErrInvalidCredentials
	}

	incomes, err := uc.incomes.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar entradas: %w", err)
	}
	for _, i := range incomes {
		if err := uc.incomes.Delete(ctx, i.ID); err != nil {
			return fmt.Errorf("erro ao remover entrada: %w", err)
		}
	}

	expenses, err := uc.expenses.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar saídas: %w", err)
	}
	for _, e := range expenses {
		if err := uc.expenses.Delete(ctx, e.ID); err != nil {
			return fmt.Errorf("erro ao remover saída: %w", err)
		}
	}

	cards, err := uc.cards.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar cartões: %w", err)
	}
	for _, c := range cards {
		purchases, err := uc.purchases.ListByCard(ctx, c.ID)
		if err != nil {
			return fmt.Errorf("erro ao listar compras parceladas: %w", err)
		}
		for _, p := range purchases {
			if err := uc.purchases.Delete(ctx, p.ID); err != nil {
				return fmt.Errorf("erro ao remover compra parcelada: %w", err)
			}
		}

		subscriptions, err := uc.subscriptions.ListByCard(ctx, c.ID)
		if err != nil {
			return fmt.Errorf("erro ao listar assinaturas: %w", err)
		}
		for _, s := range subscriptions {
			if err := uc.subscriptions.Delete(ctx, s.ID); err != nil {
				return fmt.Errorf("erro ao remover assinatura: %w", err)
			}
		}

		if err := uc.cards.Delete(ctx, c.ID); err != nil {
			return fmt.Errorf("erro ao remover cartão: %w", err)
		}
	}

	assets, err := uc.investments.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar ativos: %w", err)
	}
	for _, a := range assets {
		if err := uc.investments.Delete(ctx, a.ID); err != nil {
			return fmt.Errorf("erro ao remover ativo: %w", err)
		}
	}

	if err := uc.users.Delete(ctx, userID); err != nil {
		return fmt.Errorf("erro ao remover usuário: %w", err)
	}
	return nil
}
