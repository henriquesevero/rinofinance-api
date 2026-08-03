package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	appincome "rinofinance-api/internal/application/income"
	domainaccount "rinofinance-api/internal/domain/account"
	domaincard "rinofinance-api/internal/domain/card"
	domainexpense "rinofinance-api/internal/domain/expense"
	domainincome "rinofinance-api/internal/domain/income"
	"rinofinance-api/internal/domain/monthlystatus"
	"rinofinance-api/internal/domain/shared"
)

const uncategorizedKey = "__none__"

type AnnualMonth struct {
	Index           int
	IncomeRealized  shared.Money
	IncomePlanned   shared.Money
	ExpenseRealized shared.Money
	ExpensePlanned  shared.Money
}

type AnnualCategoryTotal struct {
	ID    string
	Total shared.Money
}

type AnnualSummary struct {
	Year                      int
	Months                    []AnnualMonth
	ExpenseCategoriesRealized []AnnualCategoryTotal
	ExpenseCategoriesPlanned  []AnnualCategoryTotal
	IncomeCategoriesRealized  []AnnualCategoryTotal
	IncomeCategoriesPlanned   []AnnualCategoryTotal
}

type GetAnnualSummaryUseCase struct {
	incomes        domainincome.Repository
	expenses       domainexpense.Repository
	incomeResolver *appincome.AccountBalanceResolver
	purchases      domaincard.InstallmentPurchaseRepository
	subscriptions  domaincard.SubscriptionRepository
	accPurchases   domainaccount.PurchaseRepository
	status         monthlystatus.Repository
}

func NewGetAnnualSummaryUseCase(
	incomes domainincome.Repository,
	expenses domainexpense.Repository,
	incomeResolver *appincome.AccountBalanceResolver,
	purchases domaincard.InstallmentPurchaseRepository,
	subscriptions domaincard.SubscriptionRepository,
	accPurchases domainaccount.PurchaseRepository,
	status monthlystatus.Repository,
) *GetAnnualSummaryUseCase {
	return &GetAnnualSummaryUseCase{
		incomes:        incomes,
		expenses:       expenses,
		incomeResolver: incomeResolver,
		purchases:      purchases,
		subscriptions:  subscriptions,
		accPurchases:   accPurchases,
		status:         status,
	}
}

type cardSource struct {
	purchases     []*domaincard.InstallmentPurchase
	subscriptions []*domaincard.Subscription
}

func (uc *GetAnnualSummaryUseCase) Execute(ctx context.Context, userID uuid.UUID, year int) (*AnnualSummary, error) {
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

	cards := map[uuid.UUID]cardSource{}
	accounts := map[uuid.UUID][]*domainaccount.Purchase{}
	for _, e := range expenses {
		switch {
		case e.IsCardLinked():
			if _, ok := cards[*e.CardID]; ok {
				continue
			}
			ps, err := uc.purchases.ListByCard(ctx, *e.CardID)
			if err != nil {
				return nil, fmt.Errorf("erro ao listar parcelas do cartão: %w", err)
			}
			ss, err := uc.subscriptions.ListByCard(ctx, *e.CardID)
			if err != nil {
				return nil, fmt.Errorf("erro ao listar assinaturas do cartão: %w", err)
			}
			cards[*e.CardID] = cardSource{purchases: ps, subscriptions: ss}
		case e.IsAccountLinked():
			if _, ok := accounts[*e.AccountID]; ok {
				continue
			}
			ps, err := uc.accPurchases.ListByAccount(ctx, *e.AccountID)
			if err != nil {
				return nil, fmt.Errorf("erro ao listar compras da conta: %w", err)
			}
			accounts[*e.AccountID] = ps
		}
	}

	months := make([]AnnualMonth, 12)
	expCatReal := moneyMap{}
	expCatPlan := moneyMap{}
	incCatReal := moneyMap{}
	incCatPlan := moneyMap{}

	for i := 0; i < 12; i++ {
		reference := time.Date(year, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
		monthKey := reference.Format("2006-01")

		incStatus, err := uc.status.ByMonth(ctx, userID, monthlystatus.Income, monthKey)
		if err != nil {
			return nil, err
		}
		expStatus, err := uc.status.ByMonth(ctx, userID, monthlystatus.Expense, monthKey)
		if err != nil {
			return nil, err
		}

		for _, e := range expenses {
			switch {
			case e.IsCardLinked():
				src := cards[*e.CardID]
				total := domaincard.MonthlyTotal(reference, src.purchases, src.subscriptions)
				if err := e.SyncAmountFromCard(total); err != nil {
					return nil, err
				}
			case e.IsAccountLinked():
				total := domainaccount.MonthlyPurchasesTotal(reference, accounts[*e.AccountID])
				if err := e.SyncAmountFromAccount(total); err != nil {
					return nil, err
				}
			}
		}

		m := AnnualMonth{
			Index:           i,
			IncomeRealized:  shared.Zero,
			IncomePlanned:   shared.Zero,
			ExpenseRealized: shared.Zero,
			ExpensePlanned:  shared.Zero,
		}
		for _, inc := range incomes {
			if !inc.Active {
				continue
			}
			m.IncomePlanned = m.IncomePlanned.Add(inc.Amount)
			incCatPlan.add(inc.CategoryID, inc.Amount)
			if incStatus[inc.ID] {
				m.IncomeRealized = m.IncomeRealized.Add(inc.Amount)
				incCatReal.add(inc.CategoryID, inc.Amount)
			}
		}
		for _, e := range expenses {
			if !e.Active {
				continue
			}
			m.ExpensePlanned = m.ExpensePlanned.Add(e.Amount)
			expCatPlan.add(e.CategoryID, e.Amount)
			if expStatus[e.ID] {
				m.ExpenseRealized = m.ExpenseRealized.Add(e.Amount)
				expCatReal.add(e.CategoryID, e.Amount)
			}
		}
		months[i] = m
	}

	return &AnnualSummary{
		Year:                      year,
		Months:                    months,
		ExpenseCategoriesRealized: expCatReal.ranked(),
		ExpenseCategoriesPlanned:  expCatPlan.ranked(),
		IncomeCategoriesRealized:  incCatReal.ranked(),
		IncomeCategoriesPlanned:   incCatPlan.ranked(),
	}, nil
}

type moneyMap map[string]shared.Money

func (m moneyMap) add(categoryID *uuid.UUID, amount shared.Money) {
	key := uncategorizedKey
	if categoryID != nil {
		key = categoryID.String()
	}
	existing, ok := m[key]
	if !ok {
		existing = shared.Zero
	}
	m[key] = existing.Add(amount)
}

func (m moneyMap) ranked() []AnnualCategoryTotal {
	out := make([]AnnualCategoryTotal, 0, len(m))
	for id, total := range m {
		out = append(out, AnnualCategoryTotal{ID: id, Total: total})
	}
	sort.Slice(out, func(i, j int) bool {

		return out[j].Total.Sub(out[i].Total).IsNegative()
	})
	return out
}
