package dto

import (
	appdashboard "rinofinance-api/internal/application/dashboard"
	"rinofinance-api/internal/domain/shared"
)

// DashboardResponse is the full payload for Aba 1 (Painel Principal).
type DashboardResponse struct {
	Incomes      []IncomeResponse  `json:"incomes"`
	Expenses     []ExpenseResponse `json:"expenses"`
	TotalIncome  shared.Money      `json:"totalIncome"`
	TotalExpense shared.Money      `json:"totalExpense"`
	NetBalance   shared.Money      `json:"netBalance"`
}

// NewDashboardResponse builds a DashboardResponse from the application
// layer's MonthlySummary.
func NewDashboardResponse(s *appdashboard.MonthlySummary) DashboardResponse {
	return DashboardResponse{
		Incomes:      NewIncomeResponseList(s.Incomes),
		Expenses:     NewExpenseResponseList(s.Expenses),
		TotalIncome:  s.TotalIncome,
		TotalExpense: s.TotalExpense,
		NetBalance:   s.NetBalance,
	}
}
