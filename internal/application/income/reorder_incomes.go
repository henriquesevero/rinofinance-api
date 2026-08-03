package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
)

type ReorderIncomesUseCase struct {
	repo domainincome.Repository
}

func NewReorderIncomesUseCase(repo domainincome.Repository) *ReorderIncomesUseCase {
	return &ReorderIncomesUseCase{repo: repo}
}

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
