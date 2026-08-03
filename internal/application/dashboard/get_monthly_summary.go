package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	appexpense "rinofinance-api/internal/application/expense"
	appincome "rinofinance-api/internal/application/income"
	domainexpense "rinofinance-api/internal/domain/expense"
	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/monthlystatus"
	"rinofinance-api/internal/domain/shared"
)

type MonthlySummary struct {
	Incomes       []*domainincome.Income
	Expenses      []*domainexpense.Expense
	TotalIncome   shared.Money
	TotalExpense  shared.Money
	NetBalance    shared.Money
	ReferenceDate time.Time
}

type GetMonthlySummaryUseCase struct {
	incomes         domainincome.Repository
	expenses        domainexpense.Repository
	resolver        *appexpense.CardAmountResolver
	accountResolver *appexpense.AccountLinkResolver
	incomeResolver  *appincome.AccountBalanceResolver
	status          monthlystatus.Repository
}

func NewGetMonthlySummaryUseCase(
	incomes domainincome.Repository,
	expenses domainexpense.Repository,
	resolver *appexpense.CardAmountResolver,
	accountResolver *appexpense.AccountLinkResolver,
	incomeResolver *appincome.AccountBalanceResolver,
	status monthlystatus.Repository,
) *GetMonthlySummaryUseCase {
	return &GetMonthlySummaryUseCase{
		incomes:         incomes,
		expenses:        expenses,
		resolver:        resolver,
		accountResolver: accountResolver,
		incomeResolver:  incomeResolver,
		status:          status,
	}
}

func (uc *GetMonthlySummaryUseCase) Execute(ctx context.Context, userID uuid.UUID, reference time.Time) (*MonthlySummary, error) {
	incomes, err := uc.incomes.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar entradas: %w", err)
	}
	if err := uc.incomeResolver.ResolveAll(ctx, incomes); err != nil {
		return nil, err
	}

	expenses, err := uc.expenses.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar saídas: %w", err)
	}
	if err := uc.resolver.ResolveAll(ctx, expenses, reference); err != nil {
		return nil, err
	}
	if err := uc.accountResolver.ResolveAll(ctx, expenses, reference); err != nil {
		return nil, err
	}

	monthKey := reference.Format("2006-01")
	incStatus, err := uc.status.ByMonth(ctx, userID, monthlystatus.Income, monthKey)
	if err != nil {
		return nil, err
	}
	for _, i := range incomes {
		i.SetReceived(incStatus[i.ID])
	}
	expStatus, err := uc.status.ByMonth(ctx, userID, monthlystatus.Expense, monthKey)
	if err != nil {
		return nil, err
	}
	for _, e := range expenses {
		e.SetPaid(expStatus[e.ID])
	}

	totalIncome := shared.Zero
	for _, i := range incomes {
		totalIncome = totalIncome.Add(i.ActiveAmount())
	}

	totalExpense := shared.Zero
	for _, e := range expenses {
		totalExpense = totalExpense.Add(e.ActiveAmount())
	}

	return &MonthlySummary{
		Incomes:       incomes,
		Expenses:      expenses,
		TotalIncome:   totalIncome,
		TotalExpense:  totalExpense,
		NetBalance:    totalIncome.Sub(totalExpense),
		ReferenceDate: reference,
	}, nil
}
