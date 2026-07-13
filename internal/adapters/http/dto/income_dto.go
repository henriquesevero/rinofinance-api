package dto

import (
	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// IncomeRequest is the payload for creating/updating an income.
type IncomeRequest struct {
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	CategoryID string       `json:"categoryId"`
}

// AccountLinkedIncomeRequest is the payload for POST
// /api/incomes/account-linked: an income whose amount mirrors an account's
// balance.
type AccountLinkedIncomeRequest struct {
	Name       string    `json:"name"`
	AccountID  uuid.UUID `json:"accountId"`
	CategoryID string    `json:"categoryId"`
}

// IncomeResponse is the public representation of an Income.
type IncomeResponse struct {
	ID         uuid.UUID    `json:"id"`
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Active     bool         `json:"active"`
	Received   bool         `json:"received"`
	CategoryID *uuid.UUID   `json:"categoryId,omitempty"`
	AccountID  *uuid.UUID   `json:"accountId,omitempty"`
}

// NewIncomeResponse builds an IncomeResponse from the domain Income.
func NewIncomeResponse(i *domainincome.Income) IncomeResponse {
	return IncomeResponse{
		ID:         i.ID,
		Name:       i.Name,
		Amount:     i.Amount,
		Active:     i.Active,
		Received:   i.Received,
		CategoryID: i.CategoryID,
		AccountID:  i.AccountID,
	}
}

// NewIncomeResponseList maps a slice of domain Incomes to their responses.
func NewIncomeResponseList(incomes []*domainincome.Income) []IncomeResponse {
	out := make([]IncomeResponse, 0, len(incomes))
	for _, i := range incomes {
		out = append(out, NewIncomeResponse(i))
	}
	return out
}
