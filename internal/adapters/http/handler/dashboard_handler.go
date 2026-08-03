package handler

import (
	"net/http"
	"strconv"
	"time"

	appdashboard "rinofinance-api/internal/application/dashboard"

	"rinofinance-api/internal/adapters/http/dto"
)

type DashboardHandler struct {
	summary *appdashboard.GetMonthlySummaryUseCase
	annual  *appdashboard.GetAnnualSummaryUseCase
}

func NewDashboardHandler(
	summary *appdashboard.GetMonthlySummaryUseCase,
	annual *appdashboard.GetAnnualSummaryUseCase,
) *DashboardHandler {
	return &DashboardHandler{summary: summary, annual: annual}
}

func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}

	summary, err := h.summary.Execute(r.Context(), userID, reference)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewDashboardResponse(summary))
}

func (h *DashboardHandler) GetAnnual(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	year := time.Now().UTC().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1970 || parsed > 9999 {
			writeError(w, errBadRequest)
			return
		}
		year = parsed
	}

	summary, err := h.annual.Execute(r.Context(), userID, year)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewAnnualSummaryResponse(summary))
}
