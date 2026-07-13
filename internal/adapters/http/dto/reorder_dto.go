package dto

// ReorderRequest is the payload for the list-reordering endpoints
// (incomes, expenses, installment purchases, subscriptions): the item IDs
// in their new display order.
type ReorderRequest struct {
	IDs []string `json:"ids"`
}
