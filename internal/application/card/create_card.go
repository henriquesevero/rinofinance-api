package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

// CardDetails carries the optional visual/financial attributes of a card
// (accent color, logo, full card-art image, credit limit and due day) so
// the create/update use cases don't grow an ever-longer argument list.
type CardDetails struct {
	Color       string
	LogoURL     string
	ImageURL    string
	CreditLimit shared.Money
	DueDay      int
	ClosingDay  int
}

// CreateCardUseCase creates a new credit card for a user.
type CreateCardUseCase struct {
	repo domaincard.CardRepository
}

// NewCreateCardUseCase wires the dependencies for CreateCardUseCase.
func NewCreateCardUseCase(repo domaincard.CardRepository) *CreateCardUseCase {
	return &CreateCardUseCase{repo: repo}
}

// Execute builds and persists a new CreditCard. All fields in details are
// optional (empty color/logo/image, zero limit/due day are valid).
func (uc *CreateCardUseCase) Execute(ctx context.Context, userID uuid.UUID, name string, details CardDetails) (*domaincard.CreditCard, error) {
	c, err := domaincard.NewCreditCard(userID, name)
	if err != nil {
		return nil, err
	}
	c.SetColor(details.Color)
	c.SetLogo(details.LogoURL)
	c.SetImage(details.ImageURL)
	c.SetCreditLimit(details.CreditLimit)
	c.SetDueDay(details.DueDay)
	c.SetClosingDay(details.ClosingDay)

	// Append new cards to the end of the user's manual ordering.
	existing, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar cartões: %w", err)
	}
	c.SetPosition(len(existing))

	if err := uc.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("erro ao criar cartão: %w", err)
	}
	return c, nil
}
