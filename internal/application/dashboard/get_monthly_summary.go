// Package dashboard orchestrates the Aba 1 "Painel Principal" summary,
// combining incomes and expenses (resolving card-linked expenses to their
// live card total) into the month's net balance.
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

// MonthlySummary is the full payload for Aba 1: every income and expense
// line for the month, plus the three headline figures (soma de entradas
// ativas, soma de saídas ativas, saldo líquido).
type MonthlySummary struct {
	Incomes       []*domainincome.Income
	Expenses      []*domainexpense.Expense
	TotalIncome   shared.Money
	TotalExpense  shared.Money
	NetBalance    shared.Money
	ReferenceDate time.Time
}

// GetMonthlySummaryUseCase computes the MonthlySummary for a user and
// reference month.
type GetMonthlySummaryUseCase struct {
	incomes         domainincome.Repository
	expenses        domainexpense.Repository
	resolver        *appexpense.CardAmountResolver
	accountResolver *appexpense.AccountLinkResolver
	incomeResolver  *appincome.AccountBalanceResolver
	status          monthlystatus.Repository
}

// NewGetMonthlySummaryUseCase wires the dependencies for
// GetMonthlySummaryUseCase.
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

// Execute loads every income and expense for userID, resolves card-linked
// expense amounts against reference, and sums only the active lines into
// the headline totals — implementing the "Regra de Ativação" at the
// dashboard level.
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

	// Overlay the per-month received/paid status so the checkmarks reflect the
	// month being viewed (they reset each month) instead of a global flag.
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
