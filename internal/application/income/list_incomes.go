package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
)

// ListIncomesUseCase lists every income belonging to a user, with
// account-linked incomes' amounts refreshed to their linked account's
// live balance.
type ListIncomesUseCase struct {
	repo     domainincome.Repository
	resolver *AccountBalanceResolver
}

// NewListIncomesUseCase wires the dependencies for ListIncomesUseCase.
func NewListIncomesUseCase(repo domainincome.Repository, resolver *AccountBalanceResolver) *ListIncomesUseCase {
	return &ListIncomesUseCase{repo: repo, resolver: resolver}
}

// Execute returns all incomes for userID, active and inactive alike — the
// toggle UI needs to render both.
func (uc *ListIncomesUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*domainincome.Income, error) {
	incomes, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar entradas: %w", err)
	}
	if err := uc.resolver.ResolveAll(ctx, incomes); err != nil {
		return nil, err
	}
	return incomes, nil
}
