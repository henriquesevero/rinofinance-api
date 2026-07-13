package expense

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the output port for persisting Expense aggregates.
type Repository interface {
	Create(ctx context.Context, e *Expense) error
	FindByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Expense, error)
	// FindByCardID returns the expense(s) linked to a given credit card, so
	// the dashboard use case can sync their amount after recomputing the
	// card's monthly total.
	FindByCardID(ctx context.Context, cardID uuid.UUID) ([]*Expense, error)
	Update(ctx context.Context, e *Expense) error
	Delete(ctx context.Context, id uuid.UUID) error
}
