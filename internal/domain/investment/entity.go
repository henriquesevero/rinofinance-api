package investment

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

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

type AssetInput struct {
	Name           string
	Ticker         string
	Class          string
	Quantity       float64
	AvgPrice       shared.Money
	CurrentPrice   shared.Money
	InvestedAmount shared.Money
	CurrentBalance shared.Money
}

type Asset struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string

	Ticker string

	Class          string
	Quantity       float64
	AvgPrice       shared.Money
	CurrentPrice   shared.Money
	InvestedAmount shared.Money
	CurrentBalance shared.Money

	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

func (a *Asset) SetActive(active bool) {
	a.Active = active
	a.UpdatedAt = time.Now().UTC()
}

func (a *Asset) ActiveBalance() shared.Money {
	if !a.Active {
		return shared.Zero
	}
	return a.CurrentBalance
}

func (a *Asset) ActiveInvested() shared.Money {
	if !a.Active {
		return shared.Zero
	}
	return a.InvestedAmount
}

func TotalPatrimony(assets []*Asset) shared.Money {
	total := shared.Zero
	for _, a := range assets {
		total = total.Add(a.ActiveBalance())
	}
	return total
}

func TotalInvested(assets []*Asset) shared.Money {
	total := shared.Zero
	for _, a := range assets {
		total = total.Add(a.ActiveInvested())
	}
	return total
}

type Provento struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	AssetID   uuid.UUID
	Amount    shared.Money
	Date      time.Time
	CreatedAt time.Time
}

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

func TotalProventos(proventos []*Provento) shared.Money {
	total := shared.Zero
	for _, p := range proventos {
		total = total.Add(p.Amount)
	}
	return total
}
