package account

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

type Account struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string

	Color string

	ImageURL string

	Balance shared.Money

	Agency string

	AccountNumber string

	AccountType string

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *Account) SetPosition(position int) {
	a.Position = position
	a.UpdatedAt = time.Now().UTC()
}

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

func (a *Account) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	a.Name = name
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a *Account) SetColor(color string) {
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#6B7280"
	}
	a.Color = color
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) SetImage(imageURL string) {
	a.ImageURL = imageURL
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) SetBalance(balance shared.Money) {
	a.Balance = balance
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) SetAgency(agency string) {
	a.Agency = strings.TrimSpace(agency)
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) SetAccountNumber(accountNumber string) {
	a.AccountNumber = strings.TrimSpace(accountNumber)
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) SetAccountType(accountType string) {
	a.AccountType = strings.TrimSpace(accountType)
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) Debit(amount shared.Money) {
	a.Balance = a.Balance.Sub(amount)
	a.UpdatedAt = time.Now().UTC()
}

func (a *Account) Credit(amount shared.Money) {
	a.Balance = a.Balance.Add(amount)
	a.UpdatedAt = time.Now().UTC()
}

type Purchase struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Name      string
	Amount    shared.Money
	Date      time.Time

	CategoryID *uuid.UUID

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func (p *Purchase) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	p.Name = name
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Purchase) UpdateAmount(amount shared.Money) error {
	if amount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	p.Amount = amount
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Purchase) SetDate(date time.Time) error {
	if date.IsZero() {
		return ErrInvalidPurchaseDate
	}
	p.Date = date
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Purchase) SetCategory(categoryID *uuid.UUID) {
	p.CategoryID = categoryID
	p.UpdatedAt = time.Now().UTC()
}

func (p *Purchase) SetPosition(position int) {
	p.Position = position
	p.UpdatedAt = time.Now().UTC()
}

func (p *Purchase) IsInMonth(reference time.Time) bool {
	return p.Date.Year() == reference.Year() && p.Date.Month() == reference.Month()
}
