package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainaccount "rinofinance-api/internal/domain/account"
	"rinofinance-api/internal/domain/shared"
)

type AccountDetails struct {
	Color         string
	ImageURL      string
	Balance       shared.Money
	Agency        string
	AccountNumber string
	AccountType   string
}

type CreateAccountUseCase struct {
	repo domainaccount.Repository
}

func NewCreateAccountUseCase(repo domainaccount.Repository) *CreateAccountUseCase {
	return &CreateAccountUseCase{repo: repo}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, userID uuid.UUID, name string, details AccountDetails) (*domainaccount.Account, error) {
	a, err := domainaccount.NewAccount(userID, name, details.Balance)
	if err != nil {
		return nil, err
	}
	a.SetColor(details.Color)
	a.SetImage(details.ImageURL)
	a.SetAgency(details.Agency)
	a.SetAccountNumber(details.AccountNumber)
	a.SetAccountType(details.AccountType)

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

type UpdateAccountUseCase struct {
	repo domainaccount.Repository
}

func NewUpdateAccountUseCase(repo domainaccount.Repository) *UpdateAccountUseCase {
	return &UpdateAccountUseCase{repo: repo}
}

func (uc *UpdateAccountUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID, name string, details AccountDetails) (*domainaccount.Account, error) {
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
	a.SetColor(details.Color)
	a.SetImage(details.ImageURL)
	a.SetBalance(details.Balance)
	a.SetAgency(details.Agency)
	a.SetAccountNumber(details.AccountNumber)
	a.SetAccountType(details.AccountType)
	if err := uc.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("erro ao atualizar conta: %w", err)
	}
	return a, nil
}

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

type AccountOverview struct {
	Account           *domainaccount.Account
	Purchases         []*domainaccount.Purchase
	MonthlyDebitTotal shared.Money
}

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

	accountIDs := make([]uuid.UUID, len(accounts))
	for i, a := range accounts {
		accountIDs[i] = a.ID
	}
	purchases, err := uc.purchases.ListByAccounts(ctx, accountIDs)
	if err != nil {
		return nil, shared.Zero, err
	}
	purchasesByAccount := make(map[uuid.UUID][]*domainaccount.Purchase)
	for _, p := range purchases {
		purchasesByAccount[p.AccountID] = append(purchasesByAccount[p.AccountID], p)
	}

	overviews := make([]AccountOverview, 0, len(accounts))
	totalBalance := shared.Zero
	for _, a := range accounts {
		accountPurchases := purchasesByAccount[a.ID]
		overviews = append(overviews, AccountOverview{
			Account:           a,
			Purchases:         accountPurchases,
			MonthlyDebitTotal: domainaccount.MonthlyPurchasesTotal(reference, accountPurchases),
		})
		totalBalance = totalBalance.Add(a.Balance)
	}
	return overviews, totalBalance, nil
}
