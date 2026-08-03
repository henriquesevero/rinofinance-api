package dto

import (
	appdashboard "rinofinance-api/internal/application/dashboard"
	"rinofinance-api/internal/domain/shared"
)

type DashboardResponse struct {
	Incomes      []IncomeResponse  `json:"incomes"`
	Expenses     []ExpenseResponse `json:"expenses"`
	TotalIncome  shared.Money      `json:"totalIncome"`
	TotalExpense shared.Money      `json:"totalExpense"`
	NetBalance   shared.Money      `json:"netBalance"`
}

func NewDashboardResponse(s *appdashboard.MonthlySummary) DashboardResponse {
	return DashboardResponse{
		Incomes:      NewIncomeResponseList(s.Incomes),
		Expenses:     NewExpenseResponseList(s.Expenses),
		TotalIncome:  s.TotalIncome,
		TotalExpense: s.TotalExpense,
		NetBalance:   s.NetBalance,
	}
}

type AnnualMonthResponse struct {
	Index           int          `json:"index"`
	IncomeRealized  shared.Money `json:"incomeRealized"`
	IncomePlanned   shared.Money `json:"incomePlanned"`
	ExpenseRealized shared.Money `json:"expenseRealized"`
	ExpensePlanned  shared.Money `json:"expensePlanned"`
}

type AnnualCategoryResponse struct {
	ID    string       `json:"id"`
	Total shared.Money `json:"total"`
}

type AnnualSummaryResponse struct {
	Year                      int                      `json:"year"`
	Months                    []AnnualMonthResponse    `json:"months"`
	ExpenseCategoriesRealized []AnnualCategoryResponse `json:"expenseCategoriesRealized"`
	ExpenseCategoriesPlanned  []AnnualCategoryResponse `json:"expenseCategoriesPlanned"`
	IncomeCategoriesRealized  []AnnualCategoryResponse `json:"incomeCategoriesRealized"`
	IncomeCategoriesPlanned   []AnnualCategoryResponse `json:"incomeCategoriesPlanned"`
}

func NewAnnualSummaryResponse(s *appdashboard.AnnualSummary) AnnualSummaryResponse {
	months := make([]AnnualMonthResponse, 0, len(s.Months))
	for _, m := range s.Months {
		months = append(months, AnnualMonthResponse{
			Index:           m.Index,
			IncomeRealized:  m.IncomeRealized,
			IncomePlanned:   m.IncomePlanned,
			ExpenseRealized: m.ExpenseRealized,
			ExpensePlanned:  m.ExpensePlanned,
		})
	}
	return AnnualSummaryResponse{
		Year:                      s.Year,
		Months:                    months,
		ExpenseCategoriesRealized: annualCategories(s.ExpenseCategoriesRealized),
		ExpenseCategoriesPlanned:  annualCategories(s.ExpenseCategoriesPlanned),
		IncomeCategoriesRealized:  annualCategories(s.IncomeCategoriesRealized),
		IncomeCategoriesPlanned:   annualCategories(s.IncomeCategoriesPlanned),
	}
}

func annualCategories(list []appdashboard.AnnualCategoryTotal) []AnnualCategoryResponse {
	out := make([]AnnualCategoryResponse, 0, len(list))
	for _, c := range list {
		out = append(out, AnnualCategoryResponse{ID: c.ID, Total: c.Total})
	}
	return out
}
