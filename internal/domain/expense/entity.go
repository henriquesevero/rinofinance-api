package expense

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

type Expense struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Amount shared.Money
	Active bool

	Paid   bool
	CardID *uuid.UUID

	CategoryID *uuid.UUID

	AccountID *uuid.UUID

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *Expense) SetAccount(accountID *uuid.UUID) {
	e.AccountID = accountID
	e.UpdatedAt = time.Now().UTC()
}

func (e *Expense) IsAccountLinked() bool {
	return e.AccountID != nil
}

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

func (e *Expense) SyncAmountFromAccount(total shared.Money) error {
	if !e.IsAccountLinked() {
		return ErrNotAccountLinked
	}
	e.Amount = total
	e.UpdatedAt = time.Now().UTC()
	return nil
}

func (e *Expense) SetCategory(categoryID *uuid.UUID) {
	e.CategoryID = categoryID
	e.UpdatedAt = time.Now().UTC()
}

func (e *Expense) SetPosition(position int) {
	e.Position = position
	e.UpdatedAt = time.Now().UTC()
}

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

func (e *Expense) IsCardLinked() bool {
	return e.CardID != nil
}

func (e *Expense) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	e.Name = name
	e.UpdatedAt = time.Now().UTC()
	return nil
}

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

func (e *Expense) Unlink() {
	e.CardID = nil
	e.UpdatedAt = time.Now().UTC()
}

func (e *Expense) SetActive(active bool) {
	e.Active = active
	e.UpdatedAt = time.Now().UTC()
}

func (e *Expense) SetPaid(paid bool) {
	e.Paid = paid
	e.UpdatedAt = time.Now().UTC()
}

func (e *Expense) ActiveAmount() shared.Money {
	if !e.Active {
		return shared.Zero
	}
	return e.Amount
}
