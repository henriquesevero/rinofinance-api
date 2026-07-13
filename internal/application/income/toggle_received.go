package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/shared"
)

// ToggleReceivedUseCase flips an income's Received flag, marking whether
// the money has actually landed this month.
type ToggleReceivedUseCase struct {
	repo domainincome.Repository
}

// NewToggleReceivedUseCase wires the dependencies for
// ToggleReceivedUseCase.
func NewToggleReceivedUseCase(repo domainincome.Repository) *ToggleReceivedUseCase {
	return &ToggleReceivedUseCase{repo: repo}
}

// Execute loads the income, verifies ownership, and flips Received.
func (uc *ToggleReceivedUseCase) Execute(ctx context.Context, userID, incomeID uuid.UUID) (*domainincome.Income, error) {
	inc, err := uc.repo.FindByID(ctx, incomeID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar entrada: %w", err)
	}
	if inc.UserID != userID {
		return nil, shared.ErrNotFound
	}

	inc.SetReceived(!inc.Received)

	if err := uc.repo.Update(ctx, inc); err != nil {
		return nil, fmt.Errorf("erro ao atualizar entrada: %w", err)
	}
	return inc, nil
}
