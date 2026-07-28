package wishlist

import (
	"context"

	"github.com/google/uuid"
)

// SectionRepository is the output port for persisting wishlist sections.
type SectionRepository interface {
	Create(ctx context.Context, s *Section) error
	FindByID(ctx context.Context, id uuid.UUID) (*Section, error)
	ListByUser(ctx context.Context, userID uuid.UUID, kind string) ([]*Section, error)
	Update(ctx context.Context, s *Section) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ItemRepository is the output port for persisting wishlist items.
type ItemRepository interface {
	Create(ctx context.Context, i *Item) error
	FindByID(ctx context.Context, id uuid.UUID) (*Item, error)
	ListByUser(ctx context.Context, userID uuid.UUID, kind string) ([]*Item, error)
	Update(ctx context.Context, i *Item) error
	Delete(ctx context.Context, id uuid.UUID) error
}
