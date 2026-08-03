package dto

import (
	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

type IncomeRequest struct {
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	CategoryID string       `json:"categoryId"`
}

type AccountLinkedIncomeRequest struct {
	Name       string    `json:"name"`
	AccountID  uuid.UUID `json:"accountId"`
	CategoryID string    `json:"categoryId"`
}

type IncomeResponse struct {
	ID         uuid.UUID    `json:"id"`
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Active     bool         `json:"active"`
	Received   bool         `json:"received"`
	CategoryID *uuid.UUID   `json:"categoryId,omitempty"`
	AccountID  *uuid.UUID   `json:"accountId,omitempty"`
}

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

func NewIncomeResponseList(incomes []*domainincome.Income) []IncomeResponse {
	out := make([]IncomeResponse, 0, len(incomes))
	for _, i := range incomes {
		out = append(out, NewIncomeResponse(i))
	}
	return out
}
