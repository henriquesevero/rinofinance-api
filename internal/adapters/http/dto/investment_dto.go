package dto

import (
	"time"

	"github.com/google/uuid"

	appinvestment "rinofinance-api/internal/application/investment"
	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// AssetRequest is the payload for creating/updating an investment asset. The
// money totals (investedAmount and currentBalance) are computed by the client
// from quantity × price, keeping fractional-share arithmetic out of the domain.
type AssetRequest struct {
	Name           string       `json:"name"`
	Ticker         string       `json:"ticker"`
	Class          string       `json:"class"`
	Quantity       float64      `json:"quantity"`
	AvgPrice       shared.Money `json:"avgPrice"`
	CurrentPrice   shared.Money `json:"currentPrice"`
	InvestedAmount shared.Money `json:"investedAmount"`
	CurrentBalance shared.Money `json:"currentBalance"`
}

// ToInput maps the request onto the domain AssetInput.
func (r AssetRequest) ToInput() domaininvestment.AssetInput {
	return domaininvestment.AssetInput{
		Name:           r.Name,
		Ticker:         r.Ticker,
		Class:          r.Class,
		Quantity:       r.Quantity,
		AvgPrice:       r.AvgPrice,
		CurrentPrice:   r.CurrentPrice,
		InvestedAmount: r.InvestedAmount,
		CurrentBalance: r.CurrentBalance,
	}
}

// AssetResponse is the public representation of an investment Asset.
type AssetResponse struct {
	ID             uuid.UUID    `json:"id"`
	Name           string       `json:"name"`
	Ticker         string       `json:"ticker"`
	Class          string       `json:"class"`
	Quantity       float64      `json:"quantity"`
	AvgPrice       shared.Money `json:"avgPrice"`
	CurrentPrice   shared.Money `json:"currentPrice"`
	InvestedAmount shared.Money `json:"investedAmount"`
	CurrentBalance shared.Money `json:"currentBalance"`
	Active         bool         `json:"active"`
}

// NewAssetResponse builds an AssetResponse from the domain Asset.
func NewAssetResponse(a *domaininvestment.Asset) AssetResponse {
	return AssetResponse{
		ID:             a.ID,
		Name:           a.Name,
		Ticker:         a.Ticker,
		Class:          a.Class,
		Quantity:       a.Quantity,
		AvgPrice:       a.AvgPrice,
		CurrentPrice:   a.CurrentPrice,
		InvestedAmount: a.InvestedAmount,
		CurrentBalance: a.CurrentBalance,
		Active:         a.Active,
	}
}

// ProventoRequest is the payload for recording a provento.
type ProventoRequest struct {
	AssetID uuid.UUID    `json:"assetId"`
	Amount  shared.Money `json:"amount"`
	Date    time.Time    `json:"date"`
}

// ProventoResponse is the public representation of a Provento.
type ProventoResponse struct {
	ID      uuid.UUID    `json:"id"`
	AssetID uuid.UUID    `json:"assetId"`
	Amount  shared.Money `json:"amount"`
	Date    time.Time    `json:"date"`
}

// NewProventoResponse builds a ProventoResponse from the domain Provento.
func NewProventoResponse(p *domaininvestment.Provento) ProventoResponse {
	return ProventoResponse{ID: p.ID, AssetID: p.AssetID, Amount: p.Amount, Date: p.Date}
}

// AssetsOverviewResponse is the full payload for GET /api/investments: every
// asset, its proventos and the portfolio totals.
type AssetsOverviewResponse struct {
	Assets         []AssetResponse    `json:"assets"`
	Proventos      []ProventoResponse `json:"proventos"`
	TotalPatrimony shared.Money       `json:"totalPatrimony"`
	TotalInvested  shared.Money       `json:"totalInvested"`
	TotalProventos shared.Money       `json:"totalProventos"`
}

// NewAssetsOverviewResponse builds the full Aba 3 payload.
func NewAssetsOverviewResponse(o appinvestment.PortfolioOverview) AssetsOverviewResponse {
	assets := make([]AssetResponse, 0, len(o.Assets))
	for _, a := range o.Assets {
		assets = append(assets, NewAssetResponse(a))
	}
	proventos := make([]ProventoResponse, 0, len(o.Proventos))
	for _, p := range o.Proventos {
		proventos = append(proventos, NewProventoResponse(p))
	}
	return AssetsOverviewResponse{
		Assets:         assets,
		Proventos:      proventos,
		TotalPatrimony: o.TotalPatrimony,
		TotalInvested:  o.TotalInvested,
		TotalProventos: o.TotalProventos,
	}
}
