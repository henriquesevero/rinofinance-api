package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the output port the application layer uses to persist and
// retrieve User aggregates. The MongoDB adapter implements this interface;
// use cases depend only on this interface, never on the concrete adapter.
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
