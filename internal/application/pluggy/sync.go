// Package pluggy orchestrates syncing a Pluggy Open Finance connection into
// the user's accounts: it mirrors checking accounts and imports their
// transactions as dated debit/credit purchases, categorized from Pluggy's
// own categories.
package pluggy

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	adapterpluggy "rinofinance-api/internal/adapters/pluggy"
	domainaccount "rinofinance-api/internal/domain/account"
	domaincategory "rinofinance-api/internal/domain/category"
	"rinofinance-api/internal/domain/shared"
)

// ErrNotConfigured is returned when Pluggy credentials weren't supplied.
var ErrNotConfigured = errors.New("integração com o Pluggy não está configurada")

// Provider is the slice of the Pluggy API this use case needs; satisfied by
// *adapters/pluggy.Client.
type Provider interface {
	Configured() bool
	GetItem(ctx context.Context, itemID string) (*adapterpluggy.Item, error)
	ListAccounts(ctx context.Context, itemID string) ([]adapterpluggy.Account, error)
	ListTransactions(ctx context.Context, accountID string) ([]adapterpluggy.Transaction, error)
}

// SyncResult summarizes what a sync produced.
type SyncResult struct {
	AccountsSynced       int
	TransactionsImported int
	TransactionsSkipped  int
}

// SyncItemUseCase mirrors a Pluggy connection's checking account(s) and their
// transactions into the user's accounts, updating in place on re-sync.
type SyncItemUseCase struct {
	provider   Provider
	accounts   domainaccount.Repository
	purchases  domainaccount.PurchaseRepository
	categories domaincategory.Repository
}

func NewSyncItemUseCase(p Provider, a domainaccount.Repository, pr domainaccount.PurchaseRepository, c domaincategory.Repository) *SyncItemUseCase {
	return &SyncItemUseCase{provider: p, accounts: a, purchases: pr, categories: c}
}

// Execute syncs every BANK account of the given Pluggy item for the user.
func (uc *SyncItemUseCase) Execute(ctx context.Context, userID uuid.UUID, itemID string) (SyncResult, error) {
	var res SyncResult
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return res, fmt.Errorf("itemId é obrigatório")
	}
	if !uc.provider.Configured() {
		return res, ErrNotConfigured
	}

	institution, color := "", ""
	if item, err := uc.provider.GetItem(ctx, itemID); err != nil {
		return res, err
	} else if item != nil {
		institution = strings.TrimSpace(item.Connector.Name)
		color = normalizeColor(item.Connector.PrimaryColor)
	}

	pAccounts, err := uc.provider.ListAccounts(ctx, itemID)
	if err != nil {
		return res, err
	}

	existing, err := uc.accounts.ListByUser(ctx, userID)
	if err != nil {
		return res, fmt.Errorf("erro ao listar contas: %w", err)
	}
	byPluggyID := make(map[string]*domainaccount.Account)
	for _, a := range existing {
		if a.PluggyAccountID != "" {
			byPluggyID[a.PluggyAccountID] = a
		}
	}
	nextPosition := len(existing)

	resolver, err := newCategoryResolver(ctx, uc.categories, userID)
	if err != nil {
		return res, err
	}

	for _, pa := range pAccounts {
		if !strings.EqualFold(pa.Type, "BANK") {
			continue // v1: checking/bank accounts only (credit cards come later)
		}
		balance := shared.NewMoneyFromFloat(pa.Balance)
		name := accountName(institution, pa)

		acc := byPluggyID[pa.ID]
		if acc == nil {
			acc, err = domainaccount.NewAccount(userID, name, balance)
			if err != nil {
				return res, err
			}
			if color != "" {
				acc.SetColor(color)
			}
			acc.LinkPluggy(itemID, pa.ID)
			acc.SetPosition(nextPosition)
			nextPosition++
			if err := uc.accounts.Create(ctx, acc); err != nil {
				return res, fmt.Errorf("erro ao criar conta sincronizada: %w", err)
			}
		} else {
			_ = acc.Rename(name)
			if color != "" {
				acc.SetColor(color)
			}
			acc.SetBalance(balance) // Pluggy is authoritative for the balance
			acc.LinkPluggy(itemID, pa.ID)
			if err := uc.accounts.Update(ctx, acc); err != nil {
				return res, fmt.Errorf("erro ao atualizar conta sincronizada: %w", err)
			}
		}
		res.AccountsSynced++

		imported, skipped, err := uc.importTransactions(ctx, resolver, acc, pa.ID)
		if err != nil {
			return res, err
		}
		res.TransactionsImported += imported
		res.TransactionsSkipped += skipped
	}

	return res, nil
}

// importTransactions imports an account's transactions as purchases, skipping
// any already imported (matched by Pluggy transaction id). The account
// balance is left untouched — Pluggy already reflects these movements.
func (uc *SyncItemUseCase) importTransactions(ctx context.Context, resolver *categoryResolver, acc *domainaccount.Account, pluggyAccountID string) (imported, skipped int, err error) {
	txs, err := uc.provider.ListTransactions(ctx, pluggyAccountID)
	if err != nil {
		return 0, 0, err
	}
	existing, err := uc.purchases.ListByAccount(ctx, acc.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao listar transações da conta: %w", err)
	}
	seen := make(map[string]bool)
	for _, p := range existing {
		if p.ExternalID != "" {
			seen[p.ExternalID] = true
		}
	}
	position := len(existing)

	for _, t := range txs {
		if t.ID == "" || seen[t.ID] {
			skipped++
			continue
		}
		date, perr := time.Parse(time.RFC3339, t.Date)
		if perr != nil {
			skipped++
			continue
		}
		direction := domainaccount.DirectionDebit
		if strings.EqualFold(t.Type, "CREDIT") || t.Amount > 0 {
			direction = domainaccount.DirectionCredit
		}
		amount := shared.NewMoneyFromFloat(math.Abs(t.Amount))

		var categoryID *uuid.UUID
		if name, icon, col := mapCategory(t.Category); name != "" {
			categoryID, err = resolver.resolve(ctx, name, icon, col)
			if err != nil {
				return imported, skipped, err
			}
		}

		p, perr := domainaccount.NewExternalPurchase(acc.ID, t.Description, amount, date, direction, t.ID, categoryID)
		if perr != nil {
			skipped++
			continue
		}
		p.SetPosition(position)
		position++
		if err := uc.purchases.Create(ctx, p); err != nil {
			return imported, skipped, fmt.Errorf("erro ao salvar transação: %w", err)
		}
		seen[t.ID] = true
		imported++
	}
	return imported, skipped, nil
}

// accountName prefers the institution name (e.g. "Itaú") so the synced
// account shows the bank's logo; it falls back to Pluggy's account name.
func accountName(institution string, pa adapterpluggy.Account) string {
	for _, candidate := range []string{institution, pa.MarketingName, pa.Name} {
		if s := strings.TrimSpace(candidate); s != "" {
			return s
		}
	}
	return "Conta"
}

// normalizeColor ensures a hex color carries a leading "#".
func normalizeColor(c string) string {
	c = strings.TrimSpace(c)
	if c == "" {
		return ""
	}
	if !strings.HasPrefix(c, "#") {
		return "#" + c
	}
	return c
}

// categoryResolver finds-or-creates a user category by name during a sync,
// caching lookups so repeated categories aren't recreated.
type categoryResolver struct {
	repo     domaincategory.Repository
	userID   uuid.UUID
	byName   map[string]*domaincategory.Category
	position int
}

func newCategoryResolver(ctx context.Context, repo domaincategory.Repository, userID uuid.UUID) (*categoryResolver, error) {
	list, err := repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar categorias: %w", err)
	}
	byName := make(map[string]*domaincategory.Category, len(list))
	for _, c := range list {
		byName[strings.ToLower(c.Name)] = c
	}
	return &categoryResolver{repo: repo, userID: userID, byName: byName, position: len(list)}, nil
}

func (r *categoryResolver) resolve(ctx context.Context, name, icon, color string) (*uuid.UUID, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return nil, nil
	}
	if c, ok := r.byName[key]; ok {
		id := c.ID
		return &id, nil
	}
	c, err := domaincategory.NewCategory(r.userID, name, color, icon)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar categoria %q: %w", name, err)
	}
	c.SetPosition(r.position)
	r.position++
	if err := r.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("erro ao salvar categoria %q: %w", name, err)
	}
	r.byName[key] = c
	id := c.ID
	return &id, nil
}
