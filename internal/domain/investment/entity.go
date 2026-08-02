// Package investment models Aba 3 (Investimentos e Patrimônio): a portfolio
// of positions (ações, FIIs, renda fixa, ...) tracked by quantity, average
// price and current price, plus the proventos (dividends/rendimentos) each one
// pays out.
package investment

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Known asset classes, used to group the portfolio by allocation.
var assetClasses = map[string]bool{
	"acao": true, "fii": true, "renda_fixa": true, "tesouro": true,
	"cripto": true, "fundo": true, "reserva": true, "outro": true,
}

func normalizeClass(c string) string {
	c = strings.TrimSpace(strings.ToLower(c))
	if assetClasses[c] {
		return c
	}
	return "outro"
}

// AssetInput carries an asset's editable fields. The money totals (invested
// and current value) are computed by the caller from quantity × price, keeping
// fractional-share arithmetic out of the domain.
type AssetInput struct {
	Name           string
	Ticker         string
	Class          string
	Quantity       float64
	AvgPrice       shared.Money // preço médio pago por cota
	CurrentPrice   shared.Money // cotação atual por cota
	InvestedAmount shared.Money // total investido (quantidade × preço médio)
	CurrentBalance shared.Money // valor atual (quantidade × preço atual)
}

// Asset is one position in the user's portfolio.
type Asset struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	// Ticker is the optional exchange symbol (e.g. "PETR4", "HGLG11").
	Ticker string
	// Class groups the asset for allocation ("acao", "fii", "renda_fixa",
	// "tesouro", "cripto", "fundo", "reserva", "outro").
	Class          string
	Quantity       float64
	AvgPrice       shared.Money
	CurrentPrice   shared.Money
	InvestedAmount shared.Money
	CurrentBalance shared.Money
	// Active mirrors the income/expense activation rule: a disabled asset
	// keeps its stored values but is excluded from the portfolio totals.
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewAsset builds a new portfolio asset for a user, active by default.
func NewAsset(userID uuid.UUID, in AssetInput) (*Asset, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, shared.ErrEmptyName
	}
	if in.CurrentBalance.IsNegative() || in.InvestedAmount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}
	now := time.Now().UTC()
	return &Asset{
		ID:             uuid.New(),
		UserID:         userID,
		Name:           name,
		Ticker:         strings.TrimSpace(in.Ticker),
		Class:          normalizeClass(in.Class),
		Quantity:       in.Quantity,
		AvgPrice:       in.AvgPrice,
		CurrentPrice:   in.CurrentPrice,
		InvestedAmount: in.InvestedAmount,
		CurrentBalance: in.CurrentBalance,
		Active:         true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Update replaces every editable field of the asset.
func (a *Asset) Update(in AssetInput) error {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return shared.ErrEmptyName
	}
	if in.CurrentBalance.IsNegative() || in.InvestedAmount.IsNegative() {
		return shared.ErrNegativeAmount
	}
	a.Name = name
	a.Ticker = strings.TrimSpace(in.Ticker)
	a.Class = normalizeClass(in.Class)
	a.Quantity = in.Quantity
	a.AvgPrice = in.AvgPrice
	a.CurrentPrice = in.CurrentPrice
	a.InvestedAmount = in.InvestedAmount
	a.CurrentBalance = in.CurrentBalance
	a.UpdatedAt = time.Now().UTC()
	return nil
}

// SetActive toggles whether this asset counts toward the portfolio totals.
func (a *Asset) SetActive(active bool) {
	a.Active = active
	a.UpdatedAt = time.Now().UTC()
}

// ActiveBalance returns the current value when active, else zero.
func (a *Asset) ActiveBalance() shared.Money {
	if !a.Active {
		return shared.Zero
	}
	return a.CurrentBalance
}

// ActiveInvested returns the invested amount when active, else zero.
func (a *Asset) ActiveInvested() shared.Money {
	if !a.Active {
		return shared.Zero
	}
	return a.InvestedAmount
}

// TotalPatrimony sums the current value across a user's active assets.
func TotalPatrimony(assets []*Asset) shared.Money {
	total := shared.Zero
	for _, a := range assets {
		total = total.Add(a.ActiveBalance())
	}
	return total
}

// TotalInvested sums the invested amount across a user's active assets.
func TotalInvested(assets []*Asset) shared.Money {
	total := shared.Zero
	for _, a := range assets {
		total = total.Add(a.ActiveInvested())
	}
	return total
}

// Provento is a dividend / rendimento received from an asset on a given date.
type Provento struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	AssetID   uuid.UUID
	Amount    shared.Money
	Date      time.Time
	CreatedAt time.Time
}

// NewProvento builds a provento entry for one of the user's assets.
func NewProvento(userID, assetID uuid.UUID, amount shared.Money, date time.Time) (*Provento, error) {
	if amount.IsNegative() {
		return nil, shared.ErrNegativeAmount
	}
	if date.IsZero() {
		date = time.Now().UTC()
	}
	return &Provento{
		ID:        uuid.New(),
		UserID:    userID,
		AssetID:   assetID,
		Amount:    amount,
		Date:      date,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// TotalProventos sums every provento amount.
func TotalProventos(proventos []*Provento) shared.Money {
	total := shared.Zero
	for _, p := range proventos {
		total = total.Add(p.Amount)
	}
	return total
}
