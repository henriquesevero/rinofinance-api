package income

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, i *Income) error
	FindByID(ctx context.Context, id uuid.UUID) (*Income, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Income, error)
	Update(ctx context.Context, i *Income) error
	Delete(ctx context.Context, id uuid.UUID) error
}
