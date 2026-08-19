package card

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domaincard "rinofinance-api/internal/domain/card"
	"rinofinance-api/internal/domain/shared"
)

type CardDetails struct {
	Color       string
	Brand       string
	LogoURL     string
	ImageURL    string
	CreditLimit shared.Money
	DueDay      int
	ClosingDay  int
}

type CreateCardUseCase struct {
	repo domaincard.CardRepository
}

func NewCreateCardUseCase(repo domaincard.CardRepository) *CreateCardUseCase {
	return &CreateCardUseCase{repo: repo}
}

func (uc *CreateCardUseCase) Execute(ctx context.Context, userID uuid.UUID, name string, details CardDetails) (*domaincard.CreditCard, error) {
	c, err := domaincard.NewCreditCard(userID, name)
	if err != nil {
		return nil, err
	}
	c.SetColor(details.Color)
	c.SetBrand(details.Brand)
	c.SetLogo(details.LogoURL)
	c.SetImage(details.ImageURL)
	c.SetCreditLimit(details.CreditLimit)
	c.SetDueDay(details.DueDay)
	c.SetClosingDay(details.ClosingDay)

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
