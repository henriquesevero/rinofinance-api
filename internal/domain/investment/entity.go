// Package investment models Aba 3 (Investimentos e Patrimônio): named
// balances such as FGTS, Previdência Privada or Tesouro Direto.
package investment

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Asset is a named balance belonging to a user (e.g. "Tesouro Direto",
// "Poupança").
type Asset struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Name           string
	CurrentBalance shared.Money
	// Active mirrors the income/expense activation rule: a disabled asset
	// keeps its stored balance but is excluded from the total patrimony
	// sum, so the user can pause tracking an account without deleting it.
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewAsset builds a new investment/patrimony asset for a user, active by
// default so it counts toward the total immediately.
func NewAsset(userID uuid.UUID, name string, currentBalance shared.Money) (*Asset, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if currentBalance.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}

	now := time.Now().UTC()
	return &Asset{
		ID:             uuid.New(),
		UserID:         userID,
		Name:           name,
		CurrentBalance: currentBalance,
		Active:         true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Rename updates the asset's display name.
func (a *Asset) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return shared.ErrEmptyName
	}
	a.Name = name
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateBalance replaces the current balance (the user re-enters the
// updated statement value; there's no transaction ledger in this version).
func (a *Asset) UpdateBalance(balance shared.Money) error {
	if balance.IsNegative() {
		return shared.ErrNegativeAmount
	}
	a.CurrentBalance = balance
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// SetActive toggles whether this asset counts toward the total
// patrimony, without touching its stored balance.
func (a *Asset) SetActive(active bool) {
	a.Active = active
	a.UpdatedAt = time.Now().UTC()
}

// ActiveBalance returns CurrentBalance when the asset is active, or
// shared.Zero when it's disabled. Callers computing the patrimony total
// must go through this method so the activation rule can't be forgotten.
func (a *Asset) ActiveBalance() shared.Money {
	if !a.Active {
		return shared.Zero
	}
	return a.CurrentBalance
}

// TotalPatrimony sums the active balances across all of a user's assets,
// feeding the "soma total do patrimônio" card in Aba 3.
func TotalPatrimony(assets []*Asset) shared.Money {
	total := shared.Zero
	for _, a := range assets {
		total = total.Add(a.ActiveBalance())
	}
	return total
}
