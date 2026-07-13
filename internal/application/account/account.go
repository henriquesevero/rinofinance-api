// Package account orchestrates use cases for the user's bank/wallet
// accounts and their debit purchases.
package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	"rinofinance-api/internal/domain/shared"
)

// CreateAccountUseCase creates a new account for a user.
type CreateAccountUseCase struct {
	repo domainaccount.Repository
}

func NewCreateAccountUseCase(repo domainaccount.Repository) *CreateAccountUseCase {
	return &CreateAccountUseCase{repo: repo}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, userID uuid.UUID, name, color, imageURL string, balance shared.Money) (*domainaccount.Account, error) {
	a, err := domainaccount.NewAccount(userID, name, balance)
	if err != nil {
		return nil, err
	}
	a.SetColor(color)
	a.SetImage(imageURL)

	existing, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar contas: %w", err)
	}
	a.SetPosition(len(existing))

	if err := uc.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao criar conta: %w", err)
	}
	return a, nil
}

// ReorderAccountsUseCase persists a new manual ordering of a user's
// accounts.
type ReorderAccountsUseCase struct {
	repo domainaccount.Repository
}

func NewReorderAccountsUseCase(repo domainaccount.Repository) *ReorderAccountsUseCase {
	return &ReorderAccountsUseCase{repo: repo}
}

func (uc *ReorderAccountsUseCase) Execute(ctx context.Context, userID uuid.UUID, orderedIDs []uuid.UUID) error {
	owned, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar contas: %w", err)
	}
	byID := make(map[uuid.UUID]*domainaccount.Account, len(owned))
	for _, a := range owned {
		byID[a.ID] = a
	}

	position := 0
	for _, id := range orderedIDs {
		a, ok := byID[id]
		if !ok {
			continue
		}
		if a.Position != position {
			a.SetPosition(position)
			if err := uc.repo.Update(ctx, a); err != nil {
				return fmt.Errorf("erro ao reordenar conta: %w", err)
			}
		}
		position++
	}
	return nil
}

// UpdateAccountUseCase renames/recolors an account, replaces its image and
// sets its defined balance.
type UpdateAccountUseCase struct {
	repo domainaccount.Repository
}

func NewUpdateAccountUseCase(repo domainaccount.Repository) *UpdateAccountUseCase {
	return &UpdateAccountUseCase{repo: repo}
}

func (uc *UpdateAccountUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID, name, color, imageURL string, balance shared.Money) (*domainaccount.Account, error) {
	a, err := uc.repo.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if a.UserID != userID {
		return nil, shared.ErrNotFound
	}
	if err := a.Rename(name); err != nil {
		return nil, err
	}
	a.SetColor(color)
	a.SetImage(imageURL)
	a.SetBalance(balance)
	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao atualizar conta: %w", err)
	}
	return a, nil
}

// DeleteAccountUseCase removes an account and cascades to its debit
// purchases. Expenses linked to it fall back to "sem conta" in the UI.
type DeleteAccountUseCase struct {
	repo      domainaccount.Repository
	purchases domainaccount.PurchaseRepository
}

func NewDeleteAccountUseCase(repo domainaccount.Repository, purchases domainaccount.PurchaseRepository) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{repo: repo, purchases: purchases}
}

func (uc *DeleteAccountUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID) error {
	a, err := uc.repo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if a.UserID != userID {
		return shared.ErrNotFound
	}
	if err := uc.purchases.DeleteByAccount(ctx, accountID); err != nil {
		return err
	}
	if err := uc.repo.Delete(ctx, accountID); err != nil {
		return fmt.Errorf("erro ao remover conta: %w", err)
	}
	return nil
}

// AccountOverview is one account plus its debit purchases and the monthly
// debit total. The current balance lives on the Account itself (debit
// purchases already decremented it).
type AccountOverview struct {
	Account           *domainaccount.Account
	Purchases         []*domainaccount.Purchase
	MonthlyDebitTotal shared.Money
}

// ListAccountsUseCase lists every account with its purchases and the
// current month's debit total.
type ListAccountsUseCase struct {
	repo      domainaccount.Repository
	purchases domainaccount.PurchaseRepository
}

func NewListAccountsUseCase(repo domainaccount.Repository, purchases domainaccount.PurchaseRepository) *ListAccountsUseCase {
	return &ListAccountsUseCase{repo: repo, purchases: purchases}
}

func (uc *ListAccountsUseCase) Execute(ctx context.Context, userID uuid.UUID, reference time.Time) ([]AccountOverview, shared.Money, error) {
	accounts, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, shared.Zero, fmt.Errorf("erro ao listar contas: %w", err)
	}

	overviews := make([]AccountOverview, 0, len(accounts))
	totalBalance := shared.Zero
	for _, a := range accounts {
		purchases, err := uc.purchases.ListByAccount(ctx, a.ID)
		if err != nil {
			return nil, shared.Zero, err
		}
		overviews = append(overviews, AccountOverview{
			Account:           a,
			Purchases:         purchases,
			MonthlyDebitTotal: domainaccount.MonthlyPurchasesTotal(reference, purchases),
		})
		totalBalance = totalBalance.Add(a.Balance)
	}
	return overviews, totalBalance, nil
}
