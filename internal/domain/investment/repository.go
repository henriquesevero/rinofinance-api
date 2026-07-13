package investment

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the output port for persisting investment Asset entities.
type Repository interface {
	Create(ctx context.Context, a *Asset) error
	FindByID(ctx context.Context, id uuid.UUID) (*Asset, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Asset, error)
	Update(ctx context.Context, a *Asset) error
	Delete(ctx context.Context, id uuid.UUID) error
}
