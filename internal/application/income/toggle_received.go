package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/monthlystatus"
	"rinofinance-api/internal/domain/shared"
)

// ToggleReceivedUseCase flips whether an income was received in a given
// month. The flag is stored per month (not on the income itself), so it
// resets every month and reflects the month being viewed.
type ToggleReceivedUseCase struct {
	repo   domainincome.Repository
	status monthlystatus.Repository
}

// NewToggleReceivedUseCase wires the dependencies for ToggleReceivedUseCase.
func NewToggleReceivedUseCase(repo domainincome.Repository, status monthlystatus.Repository) *ToggleReceivedUseCase {
	return &ToggleReceivedUseCase{repo: repo, status: status}
}

// Execute verifies ownership, flips the month's received status, and returns
// the income carrying that month's value.
func (uc *ToggleReceivedUseCase) Execute(ctx context.Context, userID, incomeID uuid.UUID, month string) (*domainincome.Income, error) {
	inc, err := uc.repo.FindByID(ctx, incomeID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar entrada: %w", err)
	}
	if inc.UserID != userID {
		return nil, shared.ErrNotFound
	}

	current, err := uc.status.Get(ctx, userID, monthlystatus.Income, incomeID, month)
	if err != nil {
		return nil, err
	}
	if err := uc.status.Set(ctx, userID, monthlystatus.Income, incomeID, month, !current); err != nil {
		return nil, err
	}

	inc.SetReceived(!current) // reflect the month's status on the returned entity
	return inc, nil
}
