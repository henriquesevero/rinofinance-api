package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type UpdateCardUseCase struct {
	repo domaincard.CardRepository
}

func NewUpdateCardUseCase(repo domaincard.CardRepository) *UpdateCardUseCase {
	return &UpdateCardUseCase{repo: repo}
}

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
	c.SetBrand(details.Brand)
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
