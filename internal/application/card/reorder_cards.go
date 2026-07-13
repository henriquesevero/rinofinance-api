package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
)

// ReorderCardsUseCase persists a new manual ordering of a user's cards.
type ReorderCardsUseCase struct {
	repo domaincard.CardRepository
}

// NewReorderCardsUseCase wires the dependencies for ReorderCardsUseCase.
func NewReorderCardsUseCase(repo domaincard.CardRepository) *ReorderCardsUseCase {
	return &ReorderCardsUseCase{repo: repo}
}

// Execute assigns each card the position of its index in orderedIDs. Only
// cards owned by the user are touched; unknown or foreign IDs are ignored
// so a stale client list can't reorder someone else's cards.
func (uc *ReorderCardsUseCase) Execute(ctx context.Context, userID uuid.UUID, orderedIDs []uuid.UUID) error {
	owned, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("erro ao listar cartões: %w", err)
	}
	byID := make(map[uuid.UUID]*domaincard.CreditCard, len(owned))
	for _, c := range owned {
		byID[c.ID] = c
	}

	position := 0
	for _, id := range orderedIDs {
		c, ok := byID[id]
		if !ok {
			continue
		}
		if c.Position != position {
			c.SetPosition(position)
			if err := uc.repo.Update(ctx, c); err != nil {
				return fmt.Errorf("erro ao reordenar cartão: %w", err)
			}
		}
		position++
	}
	return nil
}
