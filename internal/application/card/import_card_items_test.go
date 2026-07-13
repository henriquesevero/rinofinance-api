package card

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// --- in-memory fakes ---

type fakeCardRepo struct{ card *domaincard.CreditCard }

func (f *fakeCardRepo) Create(context.Context, *domaincard.CreditCard) error { return nil }
func (f *fakeCardRepo) FindByID(_ context.Context, id uuid.UUID) (*domaincard.CreditCard, error) {
	if f.card != nil && f.card.ID == id {
		return f.card, nil
	}
	return nil, shared.ErrNotFound
}
func (f *fakeCardRepo) ListByUser(context.Context, uuid.UUID) ([]*domaincard.CreditCard, error) {
	return nil, nil
}
func (f *fakeCardRepo) Update(context.Context, *domaincard.CreditCard) error { return nil }
func (f *fakeCardRepo) Delete(context.Context, uuid.UUID) error              { return nil }

type fakePurchaseRepo struct {
	created []*domaincard.InstallmentPurchase
}

func (f *fakePurchaseRepo) Create(_ context.Context, p *domaincard.InstallmentPurchase) error {
	f.created = append(f.created, p)
	return nil
}
func (f *fakePurchaseRepo) FindByID(context.Context, uuid.UUID) (*domaincard.InstallmentPurchase, error) {
	return nil, shared.ErrNotFound
}
func (f *fakePurchaseRepo) ListByCard(context.Context, uuid.UUID) ([]*domaincard.InstallmentPurchase, error) {
	return nil, nil
}
func (f *fakePurchaseRepo) Update(context.Context, *domaincard.InstallmentPurchase) error { return nil }
func (f *fakePurchaseRepo) Delete(context.Context, uuid.UUID) error                       { return nil }

type fakeSubscriptionRepo struct{ created []*domaincard.Subscription }

func (f *fakeSubscriptionRepo) Create(_ context.Context, s *domaincard.Subscription) error {
	f.created = append(f.created, s)
	return nil
}
func (f *fakeSubscriptionRepo) FindByID(context.Context, uuid.UUID) (*domaincard.Subscription, error) {
	return nil, shared.ErrNotFound
}
func (f *fakeSubscriptionRepo) ListByCard(context.Context, uuid.UUID) ([]*domaincard.Subscription, error) {
	return nil, nil
}
func (f *fakeSubscriptionRepo) Update(context.Context, *domaincard.Subscription) error { return nil }
func (f *fakeSubscriptionRepo) Delete(context.Context, uuid.UUID) error                { return nil }

func money(t *testing.T, v string) shared.Money {
	t.Helper()
	m, err := shared.NewMoney(v)
	if err != nil {
		t.Fatalf("money %q: %v", v, err)
	}
	return m
}

func TestImportCardItems_CreatesPurchasesAndSubscriptions(t *testing.T) {
	userID := uuid.New()
	card := &domaincard.CreditCard{ID: uuid.New(), UserID: userID, Name: "Itaú"}
	cards := &fakeCardRepo{card: card}
	purchases := &fakePurchaseRepo{}
	subs := &fakeSubscriptionRepo{}
	uc := NewImportCardItemsUseCase(cards, purchases, subs)

	first := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	result, err := uc.Execute(context.Background(), userID, card.ID,
		[]ImportInstallmentInput{
			{Name: "Sympla", InstallmentAmount: money(t, "57.14"), TotalInstallments: 5, FirstInstallmentDate: first, Domain: ""},
			{Name: "Uber", InstallmentAmount: money(t, "16.98"), TotalInstallments: 1, FirstInstallmentDate: first},
		},
		[]ImportSubscriptionInput{
			{Name: "Netflix", MonthlyAmount: money(t, "72.80"), Domain: "netflix.com"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InstallmentPurchases != 2 || result.Subscriptions != 1 {
		t.Fatalf("result = %+v, want 2 purchases and 1 subscription", result)
	}
	if len(purchases.created) != 2 || len(subs.created) != 1 {
		t.Fatalf("created %d purchases, %d subs", len(purchases.created), len(subs.created))
	}
	if subs.created[0].Domain != "netflix.com" {
		t.Errorf("subscription domain = %q, want netflix.com", subs.created[0].Domain)
	}
}

func TestImportCardItems_RejectsForeignCard(t *testing.T) {
	owner := uuid.New()
	card := &domaincard.CreditCard{ID: uuid.New(), UserID: owner, Name: "Itaú"}
	uc := NewImportCardItemsUseCase(&fakeCardRepo{card: card}, &fakePurchaseRepo{}, &fakeSubscriptionRepo{})

	_, err := uc.Execute(context.Background(), uuid.New() /* different user */, card.ID, nil, nil)
	if err == nil {
		t.Fatal("expected error for a card owned by another user")
	}
}

func TestImportCardItems_AbortsBeforeWritingOnInvalidItem(t *testing.T) {
	userID := uuid.New()
	card := &domaincard.CreditCard{ID: uuid.New(), UserID: userID, Name: "Itaú"}
	purchases := &fakePurchaseRepo{}
	uc := NewImportCardItemsUseCase(&fakeCardRepo{card: card}, purchases, &fakeSubscriptionRepo{})

	first := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), userID, card.ID,
		[]ImportInstallmentInput{
			{Name: "Válida", InstallmentAmount: money(t, "10.00"), TotalInstallments: 3, FirstInstallmentDate: first},
			{Name: "", InstallmentAmount: money(t, "10.00"), TotalInstallments: 3, FirstInstallmentDate: first}, // invalid: empty name
		}, nil)
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
	if len(purchases.created) != 0 {
		t.Errorf("nothing should have been persisted, got %d", len(purchases.created))
	}
}
