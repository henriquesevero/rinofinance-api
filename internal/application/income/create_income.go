// Package income orchestrates CRUD and activation use cases for Aba 1's
// "Entradas", delegating all validation to the domain/income aggregate.
package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// CreateIncomeUseCase creates a new income line for a user.
type CreateIncomeUseCase struct {
	repo domainincome.Repository
}

// NewCreateIncomeUseCase wires the dependencies for CreateIncomeUseCase.
func NewCreateIncomeUseCase(repo domainincome.Repository) *CreateIncomeUseCase {
	return &CreateIncomeUseCase{repo: repo}
}

// Execute builds and persists a new Income.
func (uc *CreateIncomeUseCase) Execute(ctx context.Context, userID uuid.UUID, name string, amount shared.Money, categoryID *uuid.UUID) (*domainincome.Income, error) {
	inc, err := domainincome.NewIncome(userID, name, amount)
	if err != nil {
		return nil, err
	}
	inc.SetCategory(categoryID)

	existing, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao contar entradas: %w", err)
	}
	inc.SetPosition(len(existing))

	if err := uc.repo.Create(ctx, inc); err != nil {
		return nil, fmt.Errorf("erro ao criar entrada: %w", err)
	}
	return inc, nil
}
