// Package card orchestrates CRUD use cases for Aba 2: credit cards,
// installment purchases and subscriptions, plus the monthly total
// calculations that feed both that tab and the linked expenses in Aba 1.
package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// verifyCardOwnership loads the card by ID and confirms it belongs to
// userID. Installment purchases and subscriptions don't carry a UserID
// themselves, so any mutation on them must go through their parent card
// to prevent a user from touching another account's data.
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
