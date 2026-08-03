package expense

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
)

type ReorderExpensesUseCase struct {
	repo domainexpense.Repository
}

func NewReorderExpensesUseCase(repo domainexpense.Repository) *ReorderExpensesUseCase {
	return &ReorderExpensesUseCase{repo: repo}
}

func (uc *ReorderExpensesUseCase) Execute(ctx context.Context, userID uuid.UUID, orderedIDs []uuid.UUID) error {
	owned, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar saídas: %w", err)
	}
	byID := make(map[uuid.UUID]*domainexpense.Expense, len(owned))
	for _, e := range owned {
		byID[e.ID] = e
	}

	position := 0
	for _, id := range orderedIDs {
		e, ok := byID[id]
		if !ok {
			continue
		}
		if e.Position != position {
			e.SetPosition(position)
			if err := uc.repo.Update(ctx, e); err != nil {
				return fmt.Errorf("erro ao reordenar saída: %w", err)
			}
		}
		position++
	}
	return nil
}
