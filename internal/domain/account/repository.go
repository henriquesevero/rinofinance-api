package account

import (
	"context"
	"time"

	"github.com/google/uuid"

	"rinofinance-api/internal/domain/shared"
)

type Repository interface {
	Create(ctx context.Context, a *Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*Account, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	Update(ctx context.Context, a *Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PurchaseRepository interface {
	Create(ctx context.Context, p *Purchase) error
	FindByID(ctx context.Context, id uuid.UUID) (*Purchase, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID) ([]*Purchase, error)
	ListByAccounts(ctx context.Context, accountIDs []uuid.UUID) ([]*Purchase, error)
	Update(ctx context.Context, p *Purchase) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAccount(ctx context.Context, accountID uuid.UUID) error
}

func TotalPurchases(purchases []*Purchase) shared.Money {
	total := shared.Zero
	for _, p := range purchases {
		total = total.Add(p.Amount)
	}
	return total
}

func MonthlyPurchasesTotal(reference time.Time, purchases []*Purchase) shared.Money {
	total := shared.Zero
	for _, p := range purchases {
		if p.IsInMonth(reference) {
			total = total.Add(p.Amount)
		}
	}
	return total
}
