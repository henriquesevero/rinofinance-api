package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/monthlystatus"
	"rinofinance-api/internal/domain/shared"
)

type ToggleReceivedUseCase struct {
	repo   domainincome.Repository
	status monthlystatus.Repository
}

func NewToggleReceivedUseCase(repo domainincome.Repository, status monthlystatus.Repository) *ToggleReceivedUseCase {
	return &ToggleReceivedUseCase{repo: repo, status: status}
}

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

	inc.SetReceived(!current)
	return inc, nil
}
