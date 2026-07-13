package dto

import (
	"github.com/google/uuid"

	domaininvestment "rinofinance-api/internal/domain/investment"
	"rinofinance-api/internal/domain/shared"
)

// AssetRequest is the payload for creating/updating an investment asset.
type AssetRequest struct {
	Name           string       `json:"name"`
	CurrentBalance shared.Money `json:"currentBalance"`
}

// AssetResponse is the public representation of an investment Asset.
type AssetResponse struct {
	ID             uuid.UUID    `json:"id"`
	Name           string       `json:"name"`
	CurrentBalance shared.Money `json:"currentBalance"`
	Active         bool         `json:"active"`
}

// NewAssetResponse builds an AssetResponse from the domain Asset.
func NewAssetResponse(a *domaininvestment.Asset) AssetResponse {
	return AssetResponse{ID: a.ID, Name: a.Name, CurrentBalance: a.CurrentBalance, Active: a.Active}
}

// AssetsOverviewResponse is the full payload for GET /api/investments:
// every asset plus the combined total patrimony card.
type AssetsOverviewResponse struct {
	Assets         []AssetResponse `json:"assets"`
	TotalPatrimony shared.Money    `json:"totalPatrimony"`
}

// NewAssetsOverviewResponse builds the full Aba 3 payload.
func NewAssetsOverviewResponse(assets []*domaininvestment.Asset, total shared.Money) AssetsOverviewResponse {
	out := make([]AssetResponse, 0, len(assets))
	for _, a := range assets {
		out = append(out, NewAssetResponse(a))
	}
	return AssetsOverviewResponse{Assets: out, TotalPatrimony: total}
}
