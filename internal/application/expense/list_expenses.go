package expense

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
)

type ListExpensesUseCase struct {
	repo            domainexpense.Repository
	resolver        *CardAmountResolver
	accountResolver *AccountLinkResolver
}

func NewListExpensesUseCase(repo domainexpense.Repository, resolver *CardAmountResolver, accountResolver *AccountLinkResolver) *ListExpensesUseCase {
	return &ListExpensesUseCase{repo: repo, resolver: resolver, accountResolver: accountResolver}
}

func (uc *ListExpensesUseCase) Execute(ctx context.Context, userID uuid.UUID, reference time.Time) ([]*domainexpense.Expense, error) {
	expenses, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar saídas: %w", err)
	}
	if err := uc.resolver.ResolveAll(ctx, expenses, reference); err != nil {
		return nil, err
	}
	if err := uc.accountResolver.ResolveAll(ctx, expenses, reference); err != nil {
		return nil, err
	}
	return expenses, nil
}
