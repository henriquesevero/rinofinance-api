package pluggy

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	adapterpluggy "rinofinance-api/internal/adapters/pluggy"
	domainaccount "rinofinance-api/internal/domain/account"
	domaincategory "rinofinance-api/internal/domain/category"
	"rinofinance-api/internal/domain/shared"
)

// TestSyncRealAPI exercises the sync use case against the real Pluggy API
// (a sandbox item) with in-memory repositories, so it never touches the
// production database. Gated on env so normal `go test` skips it:
//
//	PLUGGY_CLIENT_ID=... PLUGGY_CLIENT_SECRET=... PLUGGY_TEST_ITEM=... \
//	  go test ./internal/application/pluggy -run TestSyncRealAPI -v
func TestSyncRealAPI(t *testing.T) {
	id := os.Getenv("PLUGGY_CLIENT_ID")
	secret := os.Getenv("PLUGGY_CLIENT_SECRET")
	itemID := os.Getenv("PLUGGY_TEST_ITEM")
	if id == "" || secret == "" || itemID == "" {
		t.Skip("set PLUGGY_CLIENT_ID/PLUGGY_CLIENT_SECRET/PLUGGY_TEST_ITEM to run")
	}

	accounts := newFakeAccounts()
	purchases := newFakePurchases()
	categories := newFakeCategories()
	uc := NewSyncItemUseCase(newRealClient(id, secret), accounts, purchases, categories)

	userID := uuid.New()
	res, err := uc.Execute(context.Background(), userID, itemID)
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	t.Logf("accounts=%d imported=%d skipped=%d", res.AccountsSynced, res.TransactionsImported, res.TransactionsSkipped)
	if res.AccountsSynced == 0 {
		t.Fatalf("expected at least one BANK account synced")
	}
	if res.TransactionsImported == 0 {
		t.Fatalf("expected transactions to be imported")
	}

	for _, a := range accounts.list {
		t.Logf("account %q balance=%s pluggy=%s", a.Name, a.Balance.String(), a.PluggyAccountID)
	}
	credits, debits := 0, 0
	for _, p := range purchases.list {
		if p.IsCredit() {
			credits++
		} else {
			debits++
		}
	}
	t.Logf("purchases: %d debits, %d credits; categories created=%d", debits, credits, len(categories.list))
	for _, c := range categories.list {
		t.Logf("category %q icon=%s color=%s", c.Name, c.Icon, c.Color)
	}
	if credits == 0 {
		t.Errorf("expected at least one credit (entrada) among transactions")
	}

	// Re-sync must be idempotent: no new transactions, all skipped.
	res2, err := uc.Execute(context.Background(), userID, itemID)
	if err != nil {
		t.Fatalf("re-sync failed: %v", err)
	}
	if res2.TransactionsImported != 0 {
		t.Errorf("re-sync imported %d transactions, want 0 (dedupe)", res2.TransactionsImported)
	}
	t.Logf("re-sync: imported=%d skipped=%d", res2.TransactionsImported, res2.TransactionsSkipped)
}

// newRealClient builds the concrete adapter client, returning it as the
// Provider port (which it satisfies).
func newRealClient(id, secret string) Provider {
	return adapterpluggy.NewClient(id, secret)
}

// --- in-memory fakes -------------------------------------------------------

type fakeAccounts struct{ list []*domainaccount.Account }

func newFakeAccounts() *fakeAccounts { return &fakeAccounts{} }

func (f *fakeAccounts) Create(_ context.Context, a *domainaccount.Account) error {
	f.list = append(f.list, a)
	return nil
}
func (f *fakeAccounts) FindByID(_ context.Context, id uuid.UUID) (*domainaccount.Account, error) {
	for _, a := range f.list {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, shared.ErrNotFound
}
func (f *fakeAccounts) ListByUser(_ context.Context, userID uuid.UUID) ([]*domainaccount.Account, error) {
	var out []*domainaccount.Account
	for _, a := range f.list {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (f *fakeAccounts) Update(_ context.Context, a *domainaccount.Account) error {
	for i, x := range f.list {
		if x.ID == a.ID {
			f.list[i] = a
			return nil
		}
	}
	return shared.ErrNotFound
}
func (f *fakeAccounts) Delete(_ context.Context, id uuid.UUID) error { return nil }

type fakePurchases struct{ list []*domainaccount.Purchase }

func newFakePurchases() *fakePurchases { return &fakePurchases{} }

func (f *fakePurchases) Create(_ context.Context, p *domainaccount.Purchase) error {
	f.list = append(f.list, p)
	return nil
}
func (f *fakePurchases) FindByID(_ context.Context, id uuid.UUID) (*domainaccount.Purchase, error) {
	return nil, shared.ErrNotFound
}
func (f *fakePurchases) ListByAccount(_ context.Context, accountID uuid.UUID) ([]*domainaccount.Purchase, error) {
	var out []*domainaccount.Purchase
	for _, p := range f.list {
		if p.AccountID == accountID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (f *fakePurchases) Update(_ context.Context, p *domainaccount.Purchase) error { return nil }
func (f *fakePurchases) Delete(_ context.Context, id uuid.UUID) error              { return nil }
func (f *fakePurchases) DeleteByAccount(_ context.Context, accountID uuid.UUID) error {
	return nil
}

type fakeCategories struct{ list []*domaincategory.Category }

func newFakeCategories() *fakeCategories { return &fakeCategories{} }

func (f *fakeCategories) Create(_ context.Context, c *domaincategory.Category) error {
	f.list = append(f.list, c)
	return nil
}
func (f *fakeCategories) FindByID(_ context.Context, id uuid.UUID) (*domaincategory.Category, error) {
	return nil, shared.ErrNotFound
}
func (f *fakeCategories) ListByUser(_ context.Context, userID uuid.UUID) ([]*domaincategory.Category, error) {
	return f.list, nil
}
func (f *fakeCategories) Update(_ context.Context, c *domaincategory.Category) error { return nil }
func (f *fakeCategories) Delete(_ context.Context, id uuid.UUID) error               { return nil }
