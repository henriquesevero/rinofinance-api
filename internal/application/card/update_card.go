package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// UpdateCardUseCase renames an existing credit card and replaces its
// accent color and/or logo.
type UpdateCardUseCase struct {
	repo domaincard.CardRepository
}

// NewUpdateCardUseCase wires the dependencies for UpdateCardUseCase.
func NewUpdateCardUseCase(repo domaincard.CardRepository) *UpdateCardUseCase {
	return &UpdateCardUseCase{repo: repo}
}

// Execute loads the card, verifies ownership, then renames it and
// replaces its color/logo/image/limit/due day.
func (uc *UpdateCardUseCase) Execute(ctx context.Context, userID, cardID uuid.UUID, name string, details CardDetails) (*domaincard.CreditCard, error) {
	c, err := uc.repo.FindByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cartão: %w", err)
	}
	if c.UserID != userID {
		return nil, shared.ErrNotFound
	}

	if err := c.Rename(name); err != nil {
		return nil, err
	}
	c.SetColor(details.Color)
	c.SetLogo(details.LogoURL)
	c.SetImage(details.ImageURL)
	c.SetCreditLimit(details.CreditLimit)
	c.SetDueDay(details.DueDay)
	c.SetClosingDay(details.ClosingDay)

	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("erro ao atualizar cartão: %w", err)
	}
	return c, nil
}
