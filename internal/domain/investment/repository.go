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

// ProventoRepository is the output port for persisting proventos (dividends /
// rendimentos) received from assets.
type ProventoRepository interface {
	Create(ctx context.Context, p *Provento) error
	FindByID(ctx context.Context, id uuid.UUID) (*Provento, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Provento, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// DeleteByAsset removes every provento tied to an asset (used when the
	// asset itself is deleted).
	DeleteByAsset(ctx context.Context, assetID uuid.UUID) error
}
