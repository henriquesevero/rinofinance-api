package account

import (
	"context"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

// Repository is the output port for persisting Account entities.
type Repository interface {
	Create(ctx context.Context, a *Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*Account, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	Update(ctx context.Context, a *Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PurchaseRepository is the output port for persisting an account's debit
// Purchase entities.
type PurchaseRepository interface {
	Create(ctx context.Context, p *Purchase) error
	FindByID(ctx context.Context, id uuid.UUID) (*Purchase, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*Purchase, error)
	Update(ctx context.Context, p *Purchase) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAccount(ctx context.Context, accountID uuid.UUID) error
}

// TotalPurchases sums the amounts of the given purchases, counting debits
// only — credits ("entradas") are money in, not spending.
func TotalPurchases(purchases []*Purchase) shared.Money {
	total := shared.Zero
	for _, p := range purchases {
		if p.IsCredit() {
			continue
		}
		total = total.Add(p.Amount)
	}
	return total
}

// MonthlyPurchasesTotal sums the amounts of the purchases that fall in the
// reference month — the figure an account-linked "Compras Débito" expense
// mirrors.
func MonthlyPurchasesTotal(reference time.Time, purchases []*Purchase) shared.Money {
	total := shared.Zero
	for _, p := range purchases {
		if p.IsCredit() {
			continue
		}
		if p.IsInMonth(reference) {
			total = total.Add(p.Amount)
		}
	}
	return total
}
