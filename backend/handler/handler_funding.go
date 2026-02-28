package handler

import (
	"net/http"

	"backend/auth"
	"backend/model"
	"backend/service"

	"github.com/shopspring/decimal"
)

// FundingHandler handles funding HTTP endpoints.
type FundingHandler struct {
	svc *service.FundingService
	hub *SSEHub
}

// NewFundingHandler creates a FundingHandler.
func NewFundingHandler(svc *service.FundingService, hub *SSEHub) *FundingHandler {
	return &FundingHandler{svc: svc, hub: hub}
}

type recordFundingRequest struct {
	UserID     *int64       `json:"user_id"`
	CategoryID *int64       `json:"category_id"`
	YearMonth  string       `json:"year_month"`
	Amount     moneyRequest `json:"amount"`
}

// RecordFunding handles PUT /api/v1/workspaces/{id}/fundings.
func (h *FundingHandler) RecordFunding(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	var body recordFundingRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(body.Amount.Amount)
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid amount")
		return
	}

	currency := body.Amount.Currency
	if currency == "" {
		currency = model.DefaultCurrency
	}

	// Default user_id to caller when not specified.
	userID := callerID
	if body.UserID != nil {
		userID = model.UserID(*body.UserID)
	}

	var categoryID *model.CategoryID
	if body.CategoryID != nil {
		catID := model.CategoryID(*body.CategoryID)
		categoryID = &catID
	}

	funding, err := h.svc.Record(r.Context(), callerID, model.WorkspaceID(wsID), service.RecordFundingInput{
		UserID:     userID,
		CategoryID: categoryID,
		YearMonth:  body.YearMonth,
		Amount:     model.NewMoney(amount, currency),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "funding_changed")
	respondJSON(w, http.StatusOK, fundingResponse(funding))
}

// ListFundings handles GET /api/v1/workspaces/{id}/fundings?from=&to=.
func (h *FundingHandler) ListFundings(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	fromMonth := r.URL.Query().Get("from")
	if fromMonth == "" {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "'from' month parameter required (YYYY-MM)")
		return
	}
	toMonth := r.URL.Query().Get("to")
	if toMonth == "" {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "'to' month parameter required (YYYY-MM)")
		return
	}

	fundings, err := h.svc.ListByRange(r.Context(), callerID, model.WorkspaceID(wsID), fromMonth, toMonth)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(fundings))
	for _, f := range fundings {
		items = append(items, fundingResponse(f))
	}
	respondJSON(w, http.StatusOK, items)
}

// DeleteFunding handles DELETE /api/v1/workspaces/{id}/fundings/{fundingId}.
func (h *FundingHandler) DeleteFunding(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	fundingID, err := parseIDParam(r, "fundingId")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid funding id")
		return
	}

	err = h.svc.Delete(r.Context(), callerID, model.WorkspaceID(wsID), model.FundingID(fundingID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "funding_changed")
	w.WriteHeader(http.StatusNoContent)
}

func fundingResponse(f model.Funding) map[string]any {
	resp := map[string]any{
		"id":           f.ID,
		"workspace_id": f.WorkspaceID,
		"user_id":      f.UserID,
		"year_month":   f.YearMonth,
		"category_id":  nil,
		"amount": map[string]any{
			"amount":   f.Amount.Amount().StringFixed(2),
			"currency": f.Amount.Currency(),
		},
		"created_at": f.CreatedAt,
		"updated_at": f.UpdatedAt,
	}
	if f.CategoryID != nil {
		resp["category_id"] = *f.CategoryID
	}
	return resp
}
