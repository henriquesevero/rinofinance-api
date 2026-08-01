package card

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// Richer fakes that support lookup + delete, needed to exercise the clear
// use case (the import-test fakes are write-only).

type lookupPurchaseRepo struct {
	items   map[uuid.UUID]*domaincard.InstallmentPurchase
	deleted []uuid.UUID
}

func (r *lookupPurchaseRepo) Create(context.Context, *domaincard.InstallmentPurchase) error {
	return nil
}
func (r *lookupPurchaseRepo) FindByID(_ context.Context, id uuid.UUID) (*domaincard.InstallmentPurchase, error) {
	if p, ok := r.items[id]; ok {
		return p, nil
	}
	return nil, shared.ErrNotFound
}
func (r *lookupPurchaseRepo) ListByCard(context.Context, uuid.UUID) ([]*domaincard.InstallmentPurchase, error) {
	return nil, nil
}
func (r *lookupPurchaseRepo) Update(context.Context, *domaincard.InstallmentPurchase) error {
	return nil
}
func (r *lookupPurchaseRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted = append(r.deleted, id)
	delete(r.items, id)
	return nil
}

type lookupSubscriptionRepo struct {
	items   map[uuid.UUID]*domaincard.Subscription
	deleted []uuid.UUID
}

func (r *lookupSubscriptionRepo) Create(context.Context, *domaincard.Subscription) error { return nil }
func (r *lookupSubscriptionRepo) FindByID(_ context.Context, id uuid.UUID) (*domaincard.Subscription, error) {
	if s, ok := r.items[id]; ok {
		return s, nil
	}
	return nil, shared.ErrNotFound
}
func (r *lookupSubscriptionRepo) ListByCard(context.Context, uuid.UUID) ([]*domaincard.Subscription, error) {
	return nil, nil
}
func (r *lookupSubscriptionRepo) Update(context.Context, *domaincard.Subscription) error { return nil }
func (r *lookupSubscriptionRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.deleted = append(r.deleted, id)
	delete(r.items, id)
	return nil
}

func TestClearCardItems_DeletesSelectedOwnItems(t *testing.T) {
	userID := uuid.New()
	card := &domaincard.CreditCard{ID: uuid.New(), UserID: userID, Name: "Itaú"}

	p1 := &domaincard.InstallmentPurchase{ID: uuid.New(), CardID: card.ID, Name: "A"}
	p2 := &domaincard.InstallmentPurchase{ID: uuid.New(), CardID: card.ID, Name: "B"}
	otherCardPurchase := &domaincard.InstallmentPurchase{ID: uuid.New(), CardID: uuid.New(), Name: "Alheia"}
	s1 := &domaincard.Subscription{ID: uuid.New(), CardID: card.ID, Name: "Netflix"}

	purchases := &lookupPurchaseRepo{items: map[uuid.UUID]*domaincard.InstallmentPurchase{
		p1.ID: p1, p2.ID: p2, otherCardPurchase.ID: otherCardPurchase,
	}}
	subs := &lookupSubscriptionRepo{items: map[uuid.UUID]*domaincard.Subscription{s1.ID: s1}}
	uc := NewClearCardItemsUseCase(&fakeCardRepo{card: card}, purchases, subs)

	result, err := uc.Execute(
		context.Background(), userID, card.ID,
		[]uuid.UUID{p1.ID, otherCardPurchase.ID, uuid.New() /* nonexistent */},
		[]uuid.UUID{s1.ID},
		ClearModeDelete, time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InstallmentPurchases != 1 {
		t.Errorf("deleted %d purchases, want 1 (only the owned, existing one)", result.InstallmentPurchases)
	}
	if result.Subscriptions != 1 {
		t.Errorf("deleted %d subscriptions, want 1", result.Subscriptions)
	}
	// The other-card purchase must survive.
	if _, ok := purchases.items[otherCardPurchase.ID]; !ok {
		t.Error("a purchase from another card was wrongly deleted")
	}
	if _, ok := purchases.items[p2.ID]; !ok {
		t.Error("an unselected purchase was deleted")
	}
}

func TestClearCardItems_RejectsForeignCard(t *testing.T) {
	owner := uuid.New()
	card := &domaincard.CreditCard{ID: uuid.New(), UserID: owner, Name: "Itaú"}
	uc := NewClearCardItemsUseCase(
		&fakeCardRepo{card: card},
		&lookupPurchaseRepo{items: map[uuid.UUID]*domaincard.InstallmentPurchase{}},
		&lookupSubscriptionRepo{items: map[uuid.UUID]*domaincard.Subscription{}},
	)

	_, err := uc.Execute(context.Background(), uuid.New(), card.ID, nil, nil, ClearModeDelete, time.Now().UTC())
	if err == nil {
		t.Fatal("expected error clearing a card owned by another user")
	}
}
