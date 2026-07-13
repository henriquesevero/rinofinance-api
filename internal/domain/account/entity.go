// Package account models the user's bank/wallet accounts (e.g. "Nubank",
// "Carteira"). Each account has a user-defined balance and a list of debit
// purchases; the available balance is that defined balance minus the sum
// of its purchases (computed in the application layer, never stored).
package account

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Account is a named, user-defined balance belonging to a user.
type Account struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	// Color is a user-chosen hex color used as the account's UI accent.
	Color string
	// ImageURL optionally holds the account's "card art" image as a data:
	// URL, shown as its visual in the accounts grid (mirrors CreditCard).
	ImageURL string
	// Balance is the account's current balance. Debit purchases decrement
	// it (and restore it when removed); the user can also set it directly
	// at any time via SetBalance.
	Balance shared.Money
	// Position is the manual sort order in the accounts list (lower first).
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SetPosition sets the manual sort order of the account in the list.
func (a *Account) SetPosition(position int) {
	a.Position = position
	a.UpdatedAt = time.Now().UTC()
}

// NewAccount builds a new account for a user with a starting balance.
func NewAccount(userID uuid.UUID, name string, balance shared.Money) (*Account, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	now := time.Now().UTC()
	return &Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Balance:   balance,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Rename updates the account's display name.
func (a *Account) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	a.Name = name
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// SetColor replaces the account's accent color.
func (a *Account) SetColor(color string) {
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#6B7280"
	}
	a.Color = color
	a.UpdatedAt = time.Now().UTC()
}

// SetImage replaces the account's card-art image (a data: URL). Empty
// clears it, falling back to a color gradient with the account's name.
func (a *Account) SetImage(imageURL string) {
	a.ImageURL = imageURL
	a.UpdatedAt = time.Now().UTC()
}

// SetBalance replaces the account's current balance (manual adjustment).
func (a *Account) SetBalance(balance shared.Money) {
	a.Balance = balance
	a.UpdatedAt = time.Now().UTC()
}

// Debit subtracts amount from the balance (a debit purchase). Balances may
// go negative so an overspend is visible rather than silently rejected.
func (a *Account) Debit(amount shared.Money) {
	a.Balance = a.Balance.Sub(amount)
	a.UpdatedAt = time.Now().UTC()
}

// Credit adds amount back to the balance (when a debit purchase is removed
// or its amount is reduced).
func (a *Account) Credit(amount shared.Money) {
	a.Balance = a.Balance.Add(amount)
	a.UpdatedAt = time.Now().UTC()
}

// Purchase is a one-off debit purchase paid from an account (a "compra
// avulsa" no débito). Its amount reduces the account's available balance
// and feeds the month's account-linked "Compras Débito" expense.
type Purchase struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Name      string
	Amount    shared.Money
	Date      time.Time
	// CategoryID optionally links the purchase to a user category.
	CategoryID *uuid.UUID
	// Position is the manual sort order within the account's list.
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPurchase builds a new debit purchase under an account.
func NewPurchase(accountID uuid.UUID, name string, amount shared.Money, date time.Time) (*Purchase, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if amount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}
	if date.IsZero() {
		return nil, ErrInvalidPurchaseDate
	}
	now := time.Now().UTC()
	return &Purchase{
		ID:        uuid.New(),
		AccountID: accountID,
		Name:      name,
		Amount:    amount,
		Date:      date,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Rename updates the purchase's display name.
func (p *Purchase) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	p.Name = name
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateAmount changes the purchase's amount.
func (p *Purchase) UpdateAmount(amount shared.Money) error {
	if amount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	p.Amount = amount
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetDate replaces the purchase date.
func (p *Purchase) SetDate(date time.Time) error {
	if date.IsZero() {
		return ErrInvalidPurchaseDate
	}
	p.Date = date
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// SetCategory links (or clears, when nil) the purchase's category.
func (p *Purchase) SetCategory(categoryID *uuid.UUID) {
	p.CategoryID = categoryID
	p.UpdatedAt = time.Now().UTC()
}

// SetPosition sets the manual sort order within the account's list.
func (p *Purchase) SetPosition(position int) {
	p.Position = position
	p.UpdatedAt = time.Now().UTC()
}

// IsInMonth reports whether the purchase falls in the same calendar month
// as reference — used to compute the account's monthly debit total.
func (p *Purchase) IsInMonth(reference time.Time) bool {
	return p.Date.Year() == reference.Year() && p.Date.Month() == reference.Month()
}
