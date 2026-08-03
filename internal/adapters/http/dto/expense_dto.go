package dto

import (
	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
	"rinofinance-api/internal/domain/shared"
)

type ExpenseRequest struct {
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	CategoryID string       `json:"categoryId"`
}

type CardLinkedExpenseRequest struct {
	Name       string    `json:"name"`
	CardID     uuid.UUID `json:"cardId"`
	CategoryID string    `json:"categoryId"`
}

type AccountLinkedExpenseRequest struct {
	Name       string    `json:"name"`
	AccountID  uuid.UUID `json:"accountId"`
	CategoryID string    `json:"categoryId"`
}

type ExpenseResponse struct {
	ID         uuid.UUID    `json:"id"`
	Name       string       `json:"name"`
	Amount     shared.Money `json:"amount"`
	Active     bool         `json:"active"`
	Paid       bool         `json:"paid"`
	CardID     *uuid.UUID   `json:"cardId,omitempty"`
	CategoryID *uuid.UUID   `json:"categoryId,omitempty"`
	AccountID  *uuid.UUID   `json:"accountId,omitempty"`
}

func NewExpenseResponse(e *domainexpense.Expense) ExpenseResponse {
	return ExpenseResponse{
		ID:         e.ID,
		Name:       e.Name,
		Amount:     e.Amount,
		Active:     e.Active,
		Paid:       e.Paid,
		CardID:     e.CardID,
		CategoryID: e.CategoryID,
		AccountID:  e.AccountID,
	}
}

func NewExpenseResponseList(expenses []*domainexpense.Expense) []ExpenseResponse {
	out := make([]ExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		out = append(out, NewExpenseResponse(e))
	}
	return out
}
