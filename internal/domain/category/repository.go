package category

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the output port for persisting Category entities.
type Repository interface {
	Create(ctx context.Context, c *Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Category, error)
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}
