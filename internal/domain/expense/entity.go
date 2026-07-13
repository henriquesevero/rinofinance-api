// Package expense models "Saídas" from Aba 1 (Painel Principal): named,
// toggleable monthly expense lines. An expense may optionally be linked to
// a credit card (domain/card), in which case its Amount is not entered
// manually but kept in sync with that card's current-month total.
package expense

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Expense is a named recurring expense line belonging to a user (e.g.
// "Aluguel", "Fatura Nubank"). CardID references a domain/card.CreditCard
// by ID only — aggregates never hold pointers to other aggregates, per
// DDD aggregate boundary rules.
type Expense struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Amount shared.Money
	Active bool
	// Paid tracks whether this expense has actually been paid this month.
	// Like income's Received flag, it's purely a status marker (drives the
	// paid check in the UI) and is independent of Active — it does not
	// affect the monthly sum.
	Paid   bool
	CardID *uuid.UUID
	// CategoryID optionally links this expense to a user category for the
	// "gastos por categoria" breakdown. Nil means uncategorized.
	CategoryID *uuid.UUID
	// AccountID optionally links this expense to the bank/wallet account it
	// AccountID optionally links this expense to a bank/wallet account, in
	// which case its Amount is not entered manually but kept in sync with
	// that account's current-month debit purchases total (the "Compras
	// Débito" rule, mirroring the card link). Nil means it isn't tied to an
	// account.
	AccountID *uuid.UUID
	// Position is the manual sort order in the list (lower comes first).
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SetAccount links (or clears, when nil) the account this expense mirrors.
func (e *Expense) SetAccount(accountID *uuid.UUID) {
	e.AccountID = accountID
	e.UpdatedAt = time.Now().UTC()
}

// IsAccountLinked reports whether this expense's amount is derived from an
// account's monthly debit purchases rather than entered manually.
func (e *Expense) IsAccountLinked() bool {
	return e.AccountID != nil
}

// NewAccountLinkedExpense builds an expense whose amount mirrors an
// account's current-month debit purchases total. The initial amount is
// zero until the first SyncAmountFromAccount call.
func NewAccountLinkedExpense(userID uuid.UUID, name string, accountID uuid.UUID) (*Expense, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	now := time.Now().UTC()
	return &Expense{
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

// SyncAmountFromAccount overwrites the amount of an account-linked expense
// with the account's freshly computed monthly debit total.
func (e *Expense) SyncAmountFromAccount(total shared.Money) error {
	if !e.IsAccountLinked() {
		return ErrNotAccountLinked
	}
	e.Amount = total
	e.UpdatedAt = time.Now().UTC()
	return nil
}

// SetCategory links (or clears, when nil) the expense's category.
func (e *Expense) SetCategory(categoryID *uuid.UUID) {
	e.CategoryID = categoryID
	e.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the manual sort order of the expense in its list.
func (e *Expense) SetPosition(position int) {
	e.Position = position
	e.UpdatedAt = time.Now().UTC()
}

// NewExpense builds a standalone expense with a manually entered amount
// (not linked to any credit card).
func NewExpense(userID uuid.UUID, name string, amount shared.Money) (*Expense, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if amount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}

	now := time.Now().UTC()
	return &Expense{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Amount:    amount,
		Active:    true,
		CardID:    nil,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewCardLinkedExpense builds an expense whose amount is not entered by the
// user but derived from a credit card's monthly total (the "Vínculo com
// Cartões" rule). The initial amount is zero until the first SyncAmountFromCard call.
func NewCardLinkedExpense(userID uuid.UUID, name string, cardID uuid.UUID) (*Expense, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}

	now := time.Now().UTC()
	return &Expense{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Amount:    shared.Zero,
		Active:    true,
		CardID:    &cardID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// IsCardLinked reports whether this expense's amount is derived from a
// credit card total rather than entered manually.
func (e *Expense) IsCardLinked() bool {
	return e.CardID != nil
}

// Rename updates the expense's display name.
func (e *Expense) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	e.Name = name
	e.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateAmount changes a manually entered amount. It is rejected for
// card-linked expenses, whose amount must only ever change through
// SyncAmountFromCard, so the two data sources can never drift apart.
func (e *Expense) UpdateAmount(amount shared.Money) error {
	if e.IsCardLinked() {
		return ErrAmountManagedByCard
	}
	if e.IsAccountLinked() {
		return ErrAmountManagedByAccount
	}
	if amount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	e.Amount = amount
	e.UpdatedAt = time.Now().UTC()
	return nil
}

// SyncAmountFromCard overwrites the amount of a card-linked expense with
// the card's freshly computed current-month total. The application-layer
// dashboard use case calls this after summing installment purchases and
// subscriptions for the linked card.
func (e *Expense) SyncAmountFromCard(total shared.Money) error {
	if !e.IsCardLinked() {
		return ErrNotCardLinked
	}
	if total.IsNegative() {
		return shared.ErrNegativeAmount
	}
	e.Amount = total
	e.UpdatedAt = time.Now().UTC()
	return nil
}

// Unlink detaches this expense from its credit card, turning it into a
// standalone expense with the last synced amount, which the user can now
// edit manually. Used when the linked card itself is deleted.
func (e *Expense) Unlink() {
	e.CardID = nil
	e.UpdatedAt = time.Now().UTC()
}

// SetActive implements the "Regra de Ativação" for expenses: toggling
// on/off without deleting the record or losing its linkage/history.
func (e *Expense) SetActive(active bool) {
	e.Active = active
	e.UpdatedAt = time.Now().UTC()
}

// SetPaid marks whether this expense has already been paid this month.
// Unlike SetActive, it doesn't affect the monthly sum.
func (e *Expense) SetPaid(paid bool) {
	e.Paid = paid
	e.UpdatedAt = time.Now().UTC()
}

// ActiveAmount returns Amount when the expense counts toward the monthly
// sum, or shared.Zero when it's disabled. Dashboard totals must always go
// through this method rather than reading Amount directly.
func (e *Expense) ActiveAmount() shared.Money {
	if !e.Active {
		return shared.Zero
	}
	return e.Amount
}
