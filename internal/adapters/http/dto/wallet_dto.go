package dto

import (
	"time"

	"github.com/google/uuid"

	appaccount "rinofinance-api/internal/application/account"
	domainaccount "rinofinance-api/internal/domain/account"
	"rinofinance-api/internal/domain/shared"
)

type AccountRequest struct {
	Name     string       `json:"name"`
	Color    string       `json:"color"`
	ImageURL string       `json:"imageUrl"`
	Balance  shared.Money `json:"balance"`
}

type AccountPurchaseRequest struct {
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Date       string       `json:"date"`
	CategoryID string       `json:"categoryId"`
}

func (r AccountPurchaseRequest) ParseDate() (time.Time, error) {
	return time.Parse(DateOnlyLayout, r.Date)
}

type AccountPurchaseResponse struct {
	ID         uuid.UUID    `json:"id"`
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Date       string       `json:"date"`
	CategoryID *uuid.UUID   `json:"categoryId,omitempty"`
}

func NewAccountPurchaseResponse(p *domainaccount.Purchase) AccountPurchaseResponse {
	return AccountPurchaseResponse{
		ID:         p.ID,
		Name:       p.Name,
		Amount:     p.Amount,
		Date:       p.Date.Format(DateOnlyLayout),
		CategoryID: p.CategoryID,
	}
}

type AccountResponse struct {
	ID                uuid.UUID                 `json:"id"`
	Name              string                    `json:"name"`
	Color             string                    `json:"color,omitempty"`
	ImageURL          string                    `json:"imageUrl,omitempty"`
	Balance           shared.Money              `json:"balance"`
	MonthlyDebitTotal shared.Money              `json:"monthlyDebitTotal"`
	Purchases         []AccountPurchaseResponse `json:"purchases"`
}

func NewAccountResponse(o appaccount.AccountOverview) AccountResponse {
	purchases := make([]AccountPurchaseResponse, 0, len(o.Purchases))
	for _, p := range o.Purchases {
		purchases = append(purchases, NewAccountPurchaseResponse(p))
	}
	return AccountResponse{
		ID:                o.Account.ID,
		Name:              o.Account.Name,
		Color:             o.Account.Color,
		ImageURL:          o.Account.ImageURL,
		Balance:           o.Account.Balance,
		MonthlyDebitTotal: o.MonthlyDebitTotal,
		Purchases:         purchases,
	}
}

func NewCreatedAccountResponse(a *domainaccount.Account) AccountResponse {
	return AccountResponse{
		ID:                a.ID,
		Name:              a.Name,
		Color:             a.Color,
		ImageURL:          a.ImageURL,
		Balance:           a.Balance,
		MonthlyDebitTotal: shared.Zero,
		Purchases:         []AccountPurchaseResponse{},
	}
}

type AccountsOverviewResponse struct {
	Accounts     []AccountResponse `json:"accounts"`
	TotalBalance shared.Money      `json:"totalBalance"`
}

func NewAccountsOverviewResponse(overviews []appaccount.AccountOverview, totalBalance shared.Money) AccountsOverviewResponse {
	out := make([]AccountResponse, 0, len(overviews))
	for _, o := range overviews {
		out = append(out, NewAccountResponse(o))
	}
	return AccountsOverviewResponse{Accounts: out, TotalBalance: totalBalance}
}
