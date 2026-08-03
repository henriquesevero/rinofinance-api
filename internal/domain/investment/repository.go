package investment

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, a *Asset) error
	FindByID(ctx context.Context, id uuid.UUID) (*Asset, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Asset, error)
	Update(ctx context.Context, a *Asset) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProventoRepository interface {
	Create(ctx context.Context, p *Provento) error
	FindByID(ctx context.Context, id uuid.UUID) (*Provento, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Provento, error)
	Delete(ctx context.Context, id uuid.UUID) error

	DeleteByAsset(ctx context.Context, assetID uuid.UUID) error
}
