package dto

import (
	"time"

	"github.com/google/uuid"

	appaccount "rinofinance-api/internal/application/account"
	domainaccount "rinofinance-api/internal/domain/account"
	"rinofinance-api/internal/domain/shared"
)

// AccountRequest is the payload for creating/updating a bank/wallet
// account (distinct from the user-profile "account" endpoints).
type AccountRequest struct {
	Name     string       `json:"name"`
	Color    string       `json:"color"`
	ImageURL string       `json:"imageUrl"`
	Balance  shared.Money `json:"balance"`
}

// AccountPurchaseRequest is the payload for creating/updating a debit
// purchase under an account.
type AccountPurchaseRequest struct {
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Date       string       `json:"date"`
	CategoryID string       `json:"categoryId"`
}

// ParseDate parses the request's date-only string.
func (r AccountPurchaseRequest) ParseDate() (time.Time, error) {
	return time.Parse(DateOnlyLayout, r.Date)
}

// AccountPurchaseResponse is the public representation of a debit purchase.
type AccountPurchaseResponse struct {
	ID         uuid.UUID    `json:"id"`
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Date       string       `json:"date"`
	CategoryID *uuid.UUID   `json:"categoryId,omitempty"`
}

// NewAccountPurchaseResponse builds a response from the domain Purchase.
func NewAccountPurchaseResponse(p *domainaccount.Purchase) AccountPurchaseResponse {
	return AccountPurchaseResponse{
		ID:         p.ID,
		Name:       p.Name,
		Amount:     p.Amount,
		Date:       p.Date.Format(DateOnlyLayout),
		CategoryID: p.CategoryID,
	}
}

// AccountResponse is one account's overview: its current balance, image,
// debit purchases and the current month's debit total.
type AccountResponse struct {
	ID                uuid.UUID                 `json:"id"`
	Name              string                    `json:"name"`
	Color             string                    `json:"color,omitempty"`
	ImageURL          string                    `json:"imageUrl,omitempty"`
	Balance           shared.Money              `json:"balance"`
	MonthlyDebitTotal shared.Money              `json:"monthlyDebitTotal"`
	Purchases         []AccountPurchaseResponse `json:"purchases"`
}

// NewAccountResponse builds an AccountResponse from an application-layer
// AccountOverview.
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

// NewCreatedAccountResponse builds a response for create/update, where the
// derived figures aren't recomputed (the client refetches the overview).
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

// AccountsOverviewResponse is the full payload for GET /api/accounts.
type AccountsOverviewResponse struct {
	Accounts     []AccountResponse `json:"accounts"`
	TotalBalance shared.Money      `json:"totalBalance"`
}

// NewAccountsOverviewResponse builds the accounts overview payload.
func NewAccountsOverviewResponse(overviews []appaccount.AccountOverview, totalBalance shared.Money) AccountsOverviewResponse {
	out := make([]AccountResponse, 0, len(overviews))
	for _, o := range overviews {
		out = append(out, NewAccountResponse(o))
	}
	return AccountsOverviewResponse{Accounts: out, TotalBalance: totalBalance}
}
