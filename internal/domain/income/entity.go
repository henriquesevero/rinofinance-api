package income

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

type Income struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Amount shared.Money
	Active bool

	Received bool

	CategoryID *uuid.UUID

	AccountID *uuid.UUID

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i *Income) SetCategory(categoryID *uuid.UUID) {
	i.CategoryID = categoryID
	i.UpdatedAt = time.Now().UTC()
}

func (i *Income) SetAccount(accountID *uuid.UUID) {
	i.AccountID = accountID
	i.UpdatedAt = time.Now().UTC()
}

func (i *Income) IsAccountLinked() bool {
	return i.AccountID != nil
}

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

func (i *Income) SyncAmountFromAccount(balance shared.Money) error {
	if !i.IsAccountLinked() {
		return ErrNotAccountLinked
	}
	i.Amount = balance
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *Income) Unlink() {
	i.AccountID = nil
	i.UpdatedAt = time.Now().UTC()
}

func (i *Income) SetPosition(position int) {
	i.Position = position
	i.UpdatedAt = time.Now().UTC()
}

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

func (i *Income) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	i.Name = name
	i.UpdatedAt = time.Now().UTC()
	return nil
}

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

func (i *Income) SetActive(active bool) {
	i.Active = active
	i.UpdatedAt = time.Now().UTC()
}

func (i *Income) SetReceived(received bool) {
	i.Received = received
	i.UpdatedAt = time.Now().UTC()
}

func (i *Income) ActiveAmount() shared.Money {
	if !i.Active {
		return shared.Zero
	}
	return i.Amount
}
