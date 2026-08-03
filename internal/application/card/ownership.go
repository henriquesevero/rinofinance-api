package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

func verifyCardOwnership(ctx context.Context, cards domaincard.CardRepository, cardID, userID uuid.UUID) error {
	c, err := cards.FindByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("erro ao buscar cartão: %w", err)
	}
	if c.UserID != userID {
		return shared.ErrNotFound
	}
	return nil
}
