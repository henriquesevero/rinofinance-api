package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Endpoint  string
	P256DH    string
	Auth      string
	CreatedAt time.Time
}

func NewPushSubscription(userID uuid.UUID, endpoint, p256dh, auth string) *PushSubscription {
	return &PushSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		Endpoint:  endpoint,
		P256DH:    p256dh,
		Auth:      auth,
		CreatedAt: time.Now().UTC(),
	}
}

type Repository interface {
	Save(ctx context.Context, sub *PushSubscription) error
	DeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context) ([]*PushSubscription, error)
}
