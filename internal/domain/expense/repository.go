package expense

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, e *Expense) error
	FindByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Expense, error)

	FindByCardID(ctx context.Context, cardID uuid.UUID) ([]*Expense, error)
	Update(ctx context.Context, e *Expense) error
	Delete(ctx context.Context, id uuid.UUID) error
}
