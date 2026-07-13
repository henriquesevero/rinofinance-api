package expense

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainexpense "rinofinance-api/internal/domain/expense"
)

// ListExpensesUseCase lists every expense belonging to a user, with
// card-linked expenses' amounts refreshed to the linked card's live
// current-month total.
type ListExpensesUseCase struct {
	repo            domainexpense.Repository
	resolver        *CardAmountResolver
	accountResolver *AccountLinkResolver
}

// NewListExpensesUseCase wires the dependencies for ListExpensesUseCase.
func NewListExpensesUseCase(repo domainexpense.Repository, resolver *CardAmountResolver, accountResolver *AccountLinkResolver) *ListExpensesUseCase {
	return &ListExpensesUseCase{repo: repo, resolver: resolver, accountResolver: accountResolver}
}

// Execute returns all expenses for userID as of the reference month, with
// card-linked and account-linked amounts resolved to their live totals.
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
