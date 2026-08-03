package income

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainincome "rinofinance-api/internal/domain/income"
)

type ListIncomesUseCase struct {
	repo     domainincome.Repository
	resolver *AccountBalanceResolver
}

func NewListIncomesUseCase(repo domainincome.Repository, resolver *AccountBalanceResolver) *ListIncomesUseCase {
	return &ListIncomesUseCase{repo: repo, resolver: resolver}
}

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
