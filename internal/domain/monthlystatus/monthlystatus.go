package monthlystatus

import (
	"context"

	"github.com/google/uuid"
)

type ItemType string

const (
	Income  ItemType = "income"
	Expense ItemType = "expense"
)

type Repository interface {
	ByMonth(ctx context.Context, userID uuid.UUID, itemType ItemType, month string) (map[uuid.UUID]bool, error)

	Get(ctx context.Context, userID uuid.UUID, itemType ItemType, itemID uuid.UUID, month string) (bool, error)

	Set(ctx context.Context, userID uuid.UUID, itemType ItemType, itemID uuid.UUID, month string, done bool) error
}
