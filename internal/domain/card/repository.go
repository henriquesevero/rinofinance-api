package card

import (
	"context"

	"github.com/google/uuid"
)

type CardRepository interface {
	Create(ctx context.Context, c *CreditCard) error
	FindByID(ctx context.Context, id uuid.UUID) (*CreditCard, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*CreditCard, error)
	Update(ctx context.Context, c *CreditCard) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type InstallmentPurchaseRepository interface {
	Create(ctx context.Context, p *InstallmentPurchase) error
	FindByID(ctx context.Context, id uuid.UUID) (*InstallmentPurchase, error)
	ListByCard(ctx context.Context, cardID uuid.UUID) ([]*InstallmentPurchase, error)
	ListByCards(ctx context.Context, cardIDs []uuid.UUID) ([]*InstallmentPurchase, error)
	Update(ctx context.Context, p *InstallmentPurchase) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, s *Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	ListByCard(ctx context.Context, cardID uuid.UUID) ([]*Subscription, error)
	ListByCards(ctx context.Context, cardIDs []uuid.UUID) ([]*Subscription, error)
	Update(ctx context.Context, s *Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
}
