package handler

import (
	"net/http"

	"backend/auth"
	"backend/model"
	"backend/service"

	"github.com/shopspring/decimal"
)

// RuleHandler handles categorization rule HTTP endpoints.
type RuleHandler struct {
	svc *service.RuleService
	hub *SSEHub
}

// NewRuleHandler creates a RuleHandler.
func NewRuleHandler(svc *service.RuleService, hub *SSEHub) *RuleHandler {
	return &RuleHandler{svc: svc, hub: hub}
}

type createRuleRequest struct {
	MatchPattern     string  `json:"match_pattern"`
	TargetCategoryID int64   `json:"target_category_id"`
	Priority         int     `json:"priority"`
	AmountMin        *string `json:"amount_min"`
	AmountMax        *string `json:"amount_max"`
	BankAccountID    *int64  `json:"bank_account_id"`
	CounterpartyIBAN *string `json:"counterparty_iban"`
}

type updateRuleRequest struct {
	MatchPattern     string  `json:"match_pattern"`
	TargetCategoryID int64   `json:"target_category_id"`
	Priority         int     `json:"priority"`
	AmountMin        *string `json:"amount_min"`
	AmountMax        *string `json:"amount_max"`
	BankAccountID    *int64  `json:"bank_account_id"`
	CounterpartyIBAN *string `json:"counterparty_iban"`
}

type toggleRuleRequest struct {
	Enabled bool `json:"enabled"`
}

// parseOptionalDecimal parses a *string into a *decimal.Decimal.
// Returns nil for nil or empty strings.
func parseOptionalDecimal(s *string) (*decimal.Decimal, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	d, err := decimal.NewFromString(*s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// CreateRule handles POST /api/v1/workspaces/{id}/rules.
func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
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

	var body createRuleRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	amountMin, err := parseOptionalDecimal(body.AmountMin)
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid amount_min")
		return
	}
	amountMax, err := parseOptionalDecimal(body.AmountMax)
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid amount_max")
		return
	}

	var bankAccountID *model.BankAccountID
	if body.BankAccountID != nil {
		id := model.BankAccountID(*body.BankAccountID)
		bankAccountID = &id
	}

	var counterpartyIBAN *string
	if body.CounterpartyIBAN != nil && *body.CounterpartyIBAN != "" {
		counterpartyIBAN = body.CounterpartyIBAN
	}

	rule, err := h.svc.Create(r.Context(), callerID, model.WorkspaceID(wsID), service.CreateRuleInput{
		MatchPattern:     body.MatchPattern,
		TargetCategoryID: model.CategoryID(body.TargetCategoryID),
		Priority:         body.Priority,
		AmountMin:        amountMin,
		AmountMax:        amountMax,
		BankAccountID:    bankAccountID,
		CounterpartyIBAN: counterpartyIBAN,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "rule_changed")
	respondJSON(w, http.StatusCreated, ruleResponse(rule))
}

// UpdateRule handles PUT /api/v1/workspaces/{id}/rules/{ruleID}.
func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
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

	ruleID, err := parseIDParam(r, "ruleID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid rule id")
		return
	}

	var body updateRuleRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	amountMin, err := parseOptionalDecimal(body.AmountMin)
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid amount_min")
		return
	}
	amountMax, err := parseOptionalDecimal(body.AmountMax)
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid amount_max")
		return
	}

	var bankAccountID *model.BankAccountID
	if body.BankAccountID != nil {
		id := model.BankAccountID(*body.BankAccountID)
		bankAccountID = &id
	}

	var counterpartyIBAN *string
	if body.CounterpartyIBAN != nil && *body.CounterpartyIBAN != "" {
		counterpartyIBAN = body.CounterpartyIBAN
	}

	rule, err := h.svc.Update(r.Context(), callerID, model.WorkspaceID(wsID), model.RuleID(ruleID), service.UpdateRuleInput{
		MatchPattern:     body.MatchPattern,
		TargetCategoryID: model.CategoryID(body.TargetCategoryID),
		Priority:         body.Priority,
		AmountMin:        amountMin,
		AmountMax:        amountMax,
		BankAccountID:    bankAccountID,
		CounterpartyIBAN: counterpartyIBAN,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "rule_changed")
	respondJSON(w, http.StatusOK, ruleResponse(rule))
}

// DeleteRule handles DELETE /api/v1/workspaces/{id}/rules/{ruleID}.
func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
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

	ruleID, err := parseIDParam(r, "ruleID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid rule id")
		return
	}

	if err := h.svc.Delete(r.Context(), callerID, model.WorkspaceID(wsID), model.RuleID(ruleID)); err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "rule_changed")
	w.WriteHeader(http.StatusNoContent)
}

// ToggleRule handles PATCH /api/v1/workspaces/{id}/rules/{ruleID}/toggle.
func (h *RuleHandler) ToggleRule(w http.ResponseWriter, r *http.Request) {
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

	ruleID, err := parseIDParam(r, "ruleID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid rule id")
		return
	}

	var body toggleRuleRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	rule, err := h.svc.Toggle(r.Context(), callerID, model.WorkspaceID(wsID), model.RuleID(ruleID), body.Enabled)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "rule_changed")
	respondJSON(w, http.StatusOK, ruleResponse(rule))
}

// ListRules handles GET /api/v1/workspaces/{id}/rules.
func (h *RuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
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

	rules, err := h.svc.ListByWorkspace(r.Context(), callerID, model.WorkspaceID(wsID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(rules))
	for _, rule := range rules {
		items = append(items, ruleResponse(rule))
	}
	respondJSON(w, http.StatusOK, items)
}

func ruleResponse(r model.CategorizationRule) map[string]any {
	resp := map[string]any{
		"id":                 r.ID,
		"scope":              r.Scope,
		"match_pattern":      r.MatchPattern,
		"target_category_id": r.TargetCategoryID,
		"priority":           r.Priority,
		"enabled":            r.Enabled,
		"created_at":         r.CreatedAt,
		"updated_at":         r.UpdatedAt,
	}
	if r.OwnerID != nil {
		resp["owner_id"] = *r.OwnerID
	}
	if r.WorkspaceID != nil {
		resp["workspace_id"] = *r.WorkspaceID
	}
	if r.AmountMin != nil {
		resp["amount_min"] = r.AmountMin.StringFixed(2)
	} else {
		resp["amount_min"] = nil
	}
	if r.AmountMax != nil {
		resp["amount_max"] = r.AmountMax.StringFixed(2)
	} else {
		resp["amount_max"] = nil
	}
	if r.BankAccountID != nil {
		resp["bank_account_id"] = *r.BankAccountID
	} else {
		resp["bank_account_id"] = nil
	}
	if r.CounterpartyIBAN != nil {
		resp["counterparty_iban"] = *r.CounterpartyIBAN
	} else {
		resp["counterparty_iban"] = nil
	}
	return resp
}
