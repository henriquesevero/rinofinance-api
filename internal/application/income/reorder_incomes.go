package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
)

// ReorderIncomesUseCase persists a new manual ordering of a user's income
// lines.
type ReorderIncomesUseCase struct {
	repo domainincome.Repository
}

// NewReorderIncomesUseCase wires the dependencies for ReorderIncomesUseCase.
func NewReorderIncomesUseCase(repo domainincome.Repository) *ReorderIncomesUseCase {
	return &ReorderIncomesUseCase{repo: repo}
}

// Execute assigns each income the position of its index in orderedIDs.
// Only incomes owned by the user are touched; unknown/foreign IDs are
// ignored so a stale client list can't reorder someone else's data.
func (uc *ReorderIncomesUseCase) Execute(ctx context.Context, userID uuid.UUID, orderedIDs []uuid.UUID) error {
	owned, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar entradas: %w", err)
	}
	byID := make(map[uuid.UUID]*domainincome.Income, len(owned))
	for _, i := range owned {
		byID[i.ID] = i
	}

	position := 0
	for _, id := range orderedIDs {
		inc, ok := byID[id]
		if !ok {
			continue
		}
		if inc.Position != position {
			inc.SetPosition(position)
			if err := uc.repo.Update(ctx, inc); err != nil {
				return fmt.Errorf("erro ao reordenar entrada: %w", err)
			}
		}
		position++
	}
	return nil
}
