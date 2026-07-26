// Package monthlystatus records, per calendar month, whether a recurring
// income was received or a recurring expense was paid. This makes those
// checkmarks reset every month instead of being a single global flag on the
// item, so viewing January shows January's state, July shows July's, etc.
package monthlystatus

import (
	"context"

	"github.com/google/uuid"
)

// ItemType distinguishes which recurring item a monthly status belongs to.
type ItemType string

const (
	Income  ItemType = "income"
	Expense ItemType = "expense"
)

// Repository persists the per-month "done" flag (received for incomes, paid
// for expenses) for recurring items. Month is a "YYYY-MM" string.
type Repository interface {
	// ByMonth returns itemID -> done for every recorded status of the given
	// type in the month for the user. Items without a record are absent and
	// treated as not-done by callers.
	ByMonth(ctx context.Context, userID uuid.UUID, itemType ItemType, month string) (map[uuid.UUID]bool, error)
	// Get returns the current done value for one item in a month (false when
	// there's no record yet).
	Get(ctx context.Context, userID uuid.UUID, itemType ItemType, itemID uuid.UUID, month string) (bool, error)
	// Set upserts one item's done value for a month.
	Set(ctx context.Context, userID uuid.UUID, itemType ItemType, itemID uuid.UUID, month string, done bool) error
}
