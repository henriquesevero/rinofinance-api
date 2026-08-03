package handler

import (
	"net/http"
	"time"

	appcard "rinofinance-api/internal/application/card"

	"rinofinance-api/internal/adapters/http/dto"
)

type CardHandler struct {
	createCard                  *appcard.CreateCardUseCase
	updateCard                  *appcard.UpdateCardUseCase
	deleteCard                  *appcard.DeleteCardUseCase
	listCards                   *appcard.ListCardsUseCase
	createInstallmentPurchase   *appcard.CreateInstallmentPurchaseUseCase
	updateInstallmentPurchase   *appcard.UpdateInstallmentPurchaseUseCase
	toggleInstallmentFlag       *appcard.ToggleInstallmentPurchaseFlagUseCase
	toggleInstallmentOwed       *appcard.ToggleInstallmentPurchaseOwedUseCase
	deleteInstallmentPurchase   *appcard.DeleteInstallmentPurchaseUseCase
	createSubscription          *appcard.CreateSubscriptionUseCase
	updateSubscription          *appcard.UpdateSubscriptionUseCase
	deleteSubscription          *appcard.DeleteSubscriptionUseCase
	importCardItems             *appcard.ImportCardItemsUseCase
	clearCardItems              *appcard.ClearCardItemsUseCase
	reorderCards                *appcard.ReorderCardsUseCase
	reorderInstallmentPurchases *appcard.ReorderInstallmentPurchasesUseCase
	reorderSubscriptions        *appcard.ReorderSubscriptionsUseCase
}

func NewCardHandler(
	createCard *appcard.CreateCardUseCase,
	updateCard *appcard.UpdateCardUseCase,
	deleteCard *appcard.DeleteCardUseCase,
	listCards *appcard.ListCardsUseCase,
	createInstallmentPurchase *appcard.CreateInstallmentPurchaseUseCase,
	updateInstallmentPurchase *appcard.UpdateInstallmentPurchaseUseCase,
	toggleInstallmentFlag *appcard.ToggleInstallmentPurchaseFlagUseCase,
	toggleInstallmentOwed *appcard.ToggleInstallmentPurchaseOwedUseCase,
	deleteInstallmentPurchase *appcard.DeleteInstallmentPurchaseUseCase,
	createSubscription *appcard.CreateSubscriptionUseCase,
	updateSubscription *appcard.UpdateSubscriptionUseCase,
	deleteSubscription *appcard.DeleteSubscriptionUseCase,
	importCardItems *appcard.ImportCardItemsUseCase,
	clearCardItems *appcard.ClearCardItemsUseCase,
	reorderCards *appcard.ReorderCardsUseCase,
	reorderInstallmentPurchases *appcard.ReorderInstallmentPurchasesUseCase,
	reorderSubscriptions *appcard.ReorderSubscriptionsUseCase,
) *CardHandler {
	return &CardHandler{
		createCard:                  createCard,
		updateCard:                  updateCard,
		deleteCard:                  deleteCard,
		listCards:                   listCards,
		createInstallmentPurchase:   createInstallmentPurchase,
		updateInstallmentPurchase:   updateInstallmentPurchase,
		toggleInstallmentFlag:       toggleInstallmentFlag,
		toggleInstallmentOwed:       toggleInstallmentOwed,
		deleteInstallmentPurchase:   deleteInstallmentPurchase,
		createSubscription:          createSubscription,
		updateSubscription:          updateSubscription,
		deleteSubscription:          deleteSubscription,
		importCardItems:             importCardItems,
		clearCardItems:              clearCardItems,
		reorderCards:                reorderCards,
		reorderInstallmentPurchases: reorderInstallmentPurchases,
		reorderSubscriptions:        reorderSubscriptions,
	}
}

func (h *CardHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}

	overviews, grandTotal, err := h.listCards.Execute(r.Context(), userID, reference)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewCardsOverviewResponse(overviews, grandTotal, reference))
}

func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.CardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	c, err := h.createCard.Execute(r.Context(), userID, req.Name, req.Details())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewCardResponse(c))
}

func (h *CardHandler) UpdateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	var req dto.CardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	c, err := h.updateCard.Execute(r.Context(), userID, cardID, req.Name, req.Details())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewCardResponse(c))
}

func (h *CardHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.deleteCard.Execute(r.Context(), userID, cardID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CardHandler) CreateInstallmentPurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "cardId")
	if !ok {
		return
	}

	var req dto.InstallmentPurchaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	firstInstallmentDate, err := req.ParseFirstInstallmentDate()
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	p, err := h.createInstallmentPurchase.Execute(r.Context(), userID, cardID, req.Name, req.InstallmentAmount, req.TotalInstallments, firstInstallmentDate, req.Domain, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewInstallmentPurchaseResponse(p, firstInstallmentDate))
}

func (h *CardHandler) UpdateInstallmentPurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	purchaseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	var req dto.InstallmentPurchaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	firstInstallmentDate, err := req.ParseFirstInstallmentDate()
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	p, err := h.updateInstallmentPurchase.Execute(r.Context(), userID, purchaseID, req.Name, req.InstallmentAmount, req.TotalInstallments, firstInstallmentDate, req.Domain, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}

	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewInstallmentPurchaseResponse(p, reference))
}

func (h *CardHandler) ToggleInstallmentPurchaseFlag(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	purchaseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	p, err := h.toggleInstallmentFlag.Execute(r.Context(), userID, purchaseID)
	if err != nil {
		writeError(w, err)
		return
	}

	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewInstallmentPurchaseResponse(p, reference))
}

func (h *CardHandler) ToggleInstallmentPurchaseOwedExclusion(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	purchaseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	p, err := h.toggleInstallmentOwed.Execute(r.Context(), userID, purchaseID)
	if err != nil {
		writeError(w, err)
		return
	}

	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewInstallmentPurchaseResponse(p, reference))
}

func (h *CardHandler) DeleteInstallmentPurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	purchaseID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.deleteInstallmentPurchase.Execute(r.Context(), userID, purchaseID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CardHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "cardId")
	if !ok {
		return
	}

	var req dto.SubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	s, err := h.createSubscription.Execute(r.Context(), userID, cardID, req.Name, req.MonthlyAmount, req.Domain, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.NewSubscriptionResponse(s))
}

func (h *CardHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	subscriptionID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	var req dto.SubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	categoryID, err := parseOptionalUUID(req.CategoryID)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	s, err := h.updateSubscription.Execute(r.Context(), userID, subscriptionID, req.Name, req.MonthlyAmount, req.Domain, categoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewSubscriptionResponse(s))
}

func (h *CardHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	subscriptionID, ok := parseUUIDPathValue(w, r, "id")
	if !ok {
		return
	}

	if err := h.deleteSubscription.Execute(r.Context(), userID, subscriptionID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CardHandler) ImportFatura(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "cardId")
	if !ok {
		return
	}

	var req dto.ImportFaturaRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	installments := make([]appcard.ImportInstallmentInput, 0, len(req.InstallmentPurchases))
	for _, item := range req.InstallmentPurchases {
		date, err := time.Parse(dto.DateOnlyLayout, item.FirstInstallmentDate)
		if err != nil {
			writeError(w, errBadRequest)
			return
		}
		categoryID, err := parseOptionalUUID(item.CategoryID)
		if err != nil {
			writeError(w, errBadRequest)
			return
		}
		installments = append(installments, appcard.ImportInstallmentInput{
			Name:                 item.Name,
			InstallmentAmount:    item.InstallmentAmount,
			TotalInstallments:    item.TotalInstallments,
			FirstInstallmentDate: date,
			Domain:               item.Domain,
			CategoryID:           categoryID,
		})
	}

	subscriptions := make([]appcard.ImportSubscriptionInput, 0, len(req.Subscriptions))
	for _, item := range req.Subscriptions {
		categoryID, err := parseOptionalUUID(item.CategoryID)
		if err != nil {
			writeError(w, errBadRequest)
			return
		}
		subscriptions = append(subscriptions, appcard.ImportSubscriptionInput{
			Name:          item.Name,
			MonthlyAmount: item.MonthlyAmount,
			Domain:        item.Domain,
			CategoryID:    categoryID,
		})
	}

	var referenceMonth time.Time
	if req.ReferenceMonth != "" {
		if m, err := time.Parse("2006-01", req.ReferenceMonth); err == nil {
			referenceMonth = m
		}
	}

	result, err := h.importCardItems.Execute(r.Context(), userID, cardID, installments, subscriptions, referenceMonth)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.ImportFaturaResponse{
		InstallmentPurchases: result.InstallmentPurchases,
		Subscriptions:        result.Subscriptions,
	})
}

func (h *CardHandler) ClearCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "cardId")
	if !ok {
		return
	}

	var req dto.ClearCardRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	purchaseIDs, err := parseUUIDList(req.InstallmentPurchaseIDs)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	subscriptionIDs, err := parseUUIDList(req.SubscriptionIDs)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	reference, err := parseReferenceMonth(r)
	if err != nil {
		writeError(w, err)
		return
	}

	mode := appcard.ClearModeEnd
	if req.Mode == string(appcard.ClearModeDelete) {
		mode = appcard.ClearModeDelete
	}

	result, err := h.clearCardItems.Execute(r.Context(), userID, cardID, purchaseIDs, subscriptionIDs, mode, reference)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ClearCardResponse{
		InstallmentPurchases: result.InstallmentPurchases,
		Subscriptions:        result.Subscriptions,
	})
}

func (h *CardHandler) ReorderCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req dto.ReorderCardsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	cardIDs, err := parseUUIDList(req.CardIDs)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}

	if err := h.reorderCards.Execute(r.Context(), userID, cardIDs); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CardHandler) ReorderInstallmentPurchases(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "cardId")
	if !ok {
		return
	}
	var req dto.ReorderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	ids, err := parseUUIDList(req.IDs)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	if err := h.reorderInstallmentPurchases.Execute(r.Context(), userID, cardID, ids); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CardHandler) ReorderSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cardID, ok := parseUUIDPathValue(w, r, "cardId")
	if !ok {
		return
	}
	var req dto.ReorderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	ids, err := parseUUIDList(req.IDs)
	if err != nil {
		writeError(w, errBadRequest)
		return
	}
	if err := h.reorderSubscriptions.Execute(r.Context(), userID, cardID, ids); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
