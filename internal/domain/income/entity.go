// Package income models "Entradas" from Aba 1 (Painel Principal): named,
// toggleable monthly income lines that feed the dashboard's net balance.
package income

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Income is a named recurring income line belonging to a user (e.g.
// "Salário", "Freelance"). When Active is false its Amount must be excluded
// from any monthly sum — enforced by the caller via ActiveAmount, never by
// mutating Amount itself, so the value is preserved for when it's
// re-enabled.
type Income struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Amount shared.Money
	Active bool
	// Received tracks whether the money has actually landed this month. It
	// is purely a status marker (drives the green row in the UI) and is
	// independent of Active — it does not affect the monthly sum.
	Received bool
	// CategoryID optionally links this income to a user category. Nil means
	// uncategorized.
	CategoryID *uuid.UUID
	// AccountID optionally links this income to a bank/wallet account, in
	// which case its Amount is not entered manually but kept in sync with
	// that account's current balance (mirroring the card-linked expense
	// rule). Nil means a standalone, manually-valued income.
	AccountID *uuid.UUID
	// Position is the manual sort order in the list (lower comes first).
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SetCategory links (or clears, when nil) the income's category.
func (i *Income) SetCategory(categoryID *uuid.UUID) {
	i.CategoryID = categoryID
	i.UpdatedAt = time.Now().UTC()
}

// SetAccount links (or clears, when nil) the account this income mirrors.
func (i *Income) SetAccount(accountID *uuid.UUID) {
	i.AccountID = accountID
	i.UpdatedAt = time.Now().UTC()
}

// IsAccountLinked reports whether this income's amount is derived from a
// bank account's balance rather than entered manually.
func (i *Income) IsAccountLinked() bool {
	return i.AccountID != nil
}

// NewAccountLinkedIncome builds an income whose amount mirrors a bank
// account's balance. The initial amount is zero until the first
// SyncAmountFromAccount call.
func NewAccountLinkedIncome(userID uuid.UUID, name string, accountID uuid.UUID) (*Income, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	now := time.Now().UTC()
	return &Income{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Amount:    shared.Zero,
		Active:    true,
		AccountID: &accountID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SyncAmountFromAccount overwrites the amount of an account-linked income
// with the account's current balance. Resolved at read time by the list
// and dashboard use cases so the two never drift.
func (i *Income) SyncAmountFromAccount(balance shared.Money) error {
	if !i.IsAccountLinked() {
		return ErrNotAccountLinked
	}
	i.Amount = balance
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// Unlink detaches this income from its account, turning it into a
// standalone income with the last synced amount. Used when the linked
// account is deleted.
func (i *Income) Unlink() {
	i.AccountID = nil
	i.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the manual sort order of the income in its list.
func (i *Income) SetPosition(position int) {
	i.Position = position
	i.UpdatedAt = time.Now().UTC()
}

// NewIncome builds a new Income, defaulting to Active so it's immediately
// counted in the dashboard summary.
func NewIncome(userID uuid.UUID, name string, amount shared.Money) (*Income, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if amount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}

	now := time.Now().UTC()
	return &Income{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Amount:    amount,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Rename updates the income's display name.
func (i *Income) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	i.Name = name
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateAmount changes a manually entered amount. It is rejected for
// account-linked incomes, whose amount must only change through
// SyncAmountFromAccount.
func (i *Income) UpdateAmount(amount shared.Money) error {
	if i.IsAccountLinked() {
		return ErrAmountManagedByAccount
	}
	if amount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	i.Amount = amount
	i.UpdatedAt = time.Now().UTC()
	return nil
}

// SetActive implements the "Regra de Ativação": toggling an income on or
// off without deleting it, so history and future re-activation are kept.
func (i *Income) SetActive(active bool) {
	i.Active = active
	i.UpdatedAt = time.Now().UTC()
}

// SetReceived marks whether this income has already been received this
// month. Unlike SetActive, it doesn't affect the monthly sum.
func (i *Income) SetReceived(received bool) {
	i.Received = received
	i.UpdatedAt = time.Now().UTC()
}

// ActiveAmount returns Amount when the income counts toward the monthly
// sum, or shared.Zero when it's disabled. Callers computing dashboard
// totals must always go through this method instead of reading Amount
// directly, so the activation rule can never be forgotten at a call site.
func (i *Income) ActiveAmount() shared.Money {
	if !i.Active {
		return shared.Zero
	}
	return i.Amount
}
