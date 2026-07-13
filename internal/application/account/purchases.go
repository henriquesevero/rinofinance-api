package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	"rinofinance-api/internal/domain/shared"
)

// CreateAccountPurchaseUseCase records a one-off debit purchase under an
// account.
type CreateAccountPurchaseUseCase struct {
	accounts  domainaccount.Repository
	purchases domainaccount.PurchaseRepository
}

func NewCreateAccountPurchaseUseCase(accounts domainaccount.Repository, purchases domainaccount.PurchaseRepository) *CreateAccountPurchaseUseCase {
	return &CreateAccountPurchaseUseCase{accounts: accounts, purchases: purchases}
}

func (uc *CreateAccountPurchaseUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID, name string, amount shared.Money, date time.Time, categoryID *uuid.UUID) (*domainaccount.Purchase, error) {
	account, err := uc.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if account.UserID != userID {
		return nil, shared.ErrNotFound
	}

	p, err := domainaccount.NewPurchase(accountID, name, amount, date)
	if err != nil {
		return nil, err
	}
	p.SetCategory(categoryID)

	existing, err := uc.purchases.ListByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar compras da conta: %w", err)
	}
	p.SetPosition(len(existing))

	if err := uc.purchases.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao criar compra da conta: %w", err)
	}

	account.Debit(amount)
	if err := uc.accounts.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("erro ao atualizar saldo da conta: %w", err)
	}
	return p, nil
}

// UpdateAccountPurchaseUseCase edits a debit purchase, verifying ownership
// through its parent account.
type UpdateAccountPurchaseUseCase struct {
	accounts  domainaccount.Repository
	purchases domainaccount.PurchaseRepository
}

func NewUpdateAccountPurchaseUseCase(accounts domainaccount.Repository, purchases domainaccount.PurchaseRepository) *UpdateAccountPurchaseUseCase {
	return &UpdateAccountPurchaseUseCase{accounts: accounts, purchases: purchases}
}

func (uc *UpdateAccountPurchaseUseCase) Execute(ctx context.Context, userID, purchaseID uuid.UUID, name string, amount shared.Money, date time.Time, categoryID *uuid.UUID) (*domainaccount.Purchase, error) {
	p, err := uc.purchases.FindByID(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar compra da conta: %w", err)
	}
	account, err := uc.accounts.FindByID(ctx, p.AccountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if account.UserID != userID {
		return nil, shared.ErrNotFound
	}

	previousAmount := p.Amount
	if err := p.Rename(name); err != nil {
		return nil, err
	}
	if err := p.UpdateAmount(amount); err != nil {
		return nil, err
	}
	if err := p.SetDate(date); err != nil {
		return nil, err
	}
	p.SetCategory(categoryID)
	if err := uc.purchases.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao atualizar compra da conta: %w", err)
	}

	// Restore the old amount and apply the new one to the balance.
	account.Credit(previousAmount)
	account.Debit(amount)
	if err := uc.accounts.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("erro ao ajustar saldo da conta: %w", err)
	}
	return p, nil
}

// DeleteAccountPurchaseUseCase removes a debit purchase.
type DeleteAccountPurchaseUseCase struct {
	accounts  domainaccount.Repository
	purchases domainaccount.PurchaseRepository
}

func NewDeleteAccountPurchaseUseCase(accounts domainaccount.Repository, purchases domainaccount.PurchaseRepository) *DeleteAccountPurchaseUseCase {
	return &DeleteAccountPurchaseUseCase{accounts: accounts, purchases: purchases}
}

func (uc *DeleteAccountPurchaseUseCase) Execute(ctx context.Context, userID, purchaseID uuid.UUID) error {
	p, err := uc.purchases.FindByID(ctx, purchaseID)
	if err != nil {
		return fmt.Errorf("erro ao buscar compra da conta: %w", err)
	}
	account, err := uc.accounts.FindByID(ctx, p.AccountID)
	if err != nil {
		return fmt.Errorf("erro ao buscar conta: %w", err)
	}
	if account.UserID != userID {
		return shared.ErrNotFound
	}
	if err := uc.purchases.Delete(ctx, purchaseID); err != nil {
		return fmt.Errorf("erro ao remover compra da conta: %w", err)
	}

	account.Credit(p.Amount)
	if err := uc.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("erro ao restaurar saldo da conta: %w", err)
	}
	return nil
}

// AccountDebitResolver refreshes an account-linked expense's in-memory
// Amount from its account's current-month debit purchases total (mirroring
// the CardAmountResolver). A pure read-time concern.
type AccountDebitResolver struct {
	purchases domainaccount.PurchaseRepository
}

func NewAccountDebitResolver(purchases domainaccount.PurchaseRepository) *AccountDebitResolver {
	return &AccountDebitResolver{purchases: purchases}
}

// MonthlyDebitTotal returns the sum of an account's purchases in the
// reference month.
func (r *AccountDebitResolver) MonthlyDebitTotal(ctx context.Context, accountID uuid.UUID, reference time.Time) (shared.Money, error) {
	purchases, err := r.purchases.ListByAccount(ctx, accountID)
	if err != nil {
		return shared.Zero, fmt.Errorf("erro ao listar compras da conta vinculada: %w", err)
	}
	return domainaccount.MonthlyPurchasesTotal(reference, purchases), nil
}
