package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"backend/auth"
	"backend/model"
	"backend/service"

	"github.com/go-chi/chi/v5"
)

// WorkspaceHandler handles workspace and member HTTP endpoints.
type WorkspaceHandler struct {
	svc *service.WorkspaceService
	hub *SSEHub
}

// NewWorkspaceHandler creates a WorkspaceHandler.
func NewWorkspaceHandler(svc *service.WorkspaceService, hub *SSEHub) *WorkspaceHandler {
	return &WorkspaceHandler{svc: svc, hub: hub}
}

// CreateWorkspace handles POST /api/v1/workspaces.
func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	ws, err := h.svc.CreateWorkspace(r.Context(), userID, service.CreateWorkspaceInput{
		Name:        body.Name,
		Description: body.Description,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(int64(ws.ID), "workspace_changed")
	respondJSON(w, http.StatusCreated, workspaceResponse(ws))
}

// ListWorkspaces handles GET /api/v1/workspaces.
func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	workspaces, err := h.svc.ListWorkspaces(r.Context(), userID)
	if err != nil {
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "failed to list workspaces")
		return
	}

	result := make([]map[string]any, 0, len(workspaces))
	for _, ws := range workspaces {
		result = append(result, workspaceResponse(ws))
	}
	respondJSON(w, http.StatusOK, result)
}

// GetWorkspace handles GET /api/v1/workspaces/{id}.
func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	ws, err := h.svc.GetWorkspace(r.Context(), userID, model.WorkspaceID(wsID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, workspaceResponse(ws))
}

// UpdateWorkspace handles PUT /api/v1/workspaces/{id}.
func (h *WorkspaceHandler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	ws, err := h.svc.UpdateWorkspace(r.Context(), userID, model.WorkspaceID(wsID), service.UpdateWorkspaceInput{
		Name:        body.Name,
		Description: body.Description,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "workspace_changed")
	respondJSON(w, http.StatusOK, workspaceResponse(ws))
}

// DeleteWorkspace handles DELETE /api/v1/workspaces/{id}.
func (h *WorkspaceHandler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	if err := h.svc.DeleteWorkspace(r.Context(), userID, model.WorkspaceID(wsID)); err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "workspace_changed")
	w.WriteHeader(http.StatusNoContent)
}

// AddMember handles POST /api/v1/workspaces/{id}/members.
func (h *WorkspaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	var body struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	member, err := h.svc.AddMember(r.Context(), userID, model.WorkspaceID(wsID), service.AddMemberInput{
		UserID: model.UserID(body.UserID),
		Role:   model.MemberRole(body.Role),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "member_changed")
	respondJSON(w, http.StatusCreated, memberResponse(member))
}

// RemoveMember handles DELETE /api/v1/workspaces/{id}/members/{userID}.
func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	wsID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid workspace id")
		return
	}

	targetUserID, err := parseIDParam(r, "userID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid user id")
		return
	}

	if err := h.svc.RemoveMember(r.Context(), userID, model.WorkspaceID(wsID), model.UserID(targetUserID)); err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "member_changed")
	w.WriteHeader(http.StatusNoContent)
}

func workspaceResponse(ws model.Workspace) map[string]any {
	return map[string]any{
		"id":           ws.ID,
		"name":         ws.Name,
		"description":  ws.Description,
		"owner_id":     ws.OwnerID,
		"member_count": ws.MemberCount,
		"created_at":   ws.CreatedAt,
		"updated_at":   ws.UpdatedAt,
	}
}

func memberResponse(m model.WorkspaceMember) map[string]any {
	return map[string]any{
		"workspace_id": m.WorkspaceID,
		"user_id":      m.UserID,
		"role":         m.Role,
		"created_at":   m.CreatedAt,
	}
}

func parseIDParam(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, param), 10, 64)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrWorkspaceNotFound),
		errors.Is(err, model.ErrCategoryNotFound),
		errors.Is(err, model.ErrFundingNotFound),
		errors.Is(err, model.ErrRuleNotFound),
		errors.Is(err, model.ErrBankConnectionNotFound),
		errors.Is(err, model.ErrBankAccountNotFound),
		errors.Is(err, model.ErrTransactionNotFound),
		errors.Is(err, model.ErrInviteNotFound),
		errors.Is(err, model.ErrUserNotFound):
		respondAppError(w, http.StatusNotFound, err)
	case errors.Is(err, model.ErrNotWorkspaceMember),
		errors.Is(err, model.ErrInsufficientPermission):
		respondAppError(w, http.StatusForbidden, err)
	case errors.Is(err, model.ErrInviteExpired),
		errors.Is(err, model.ErrInviteUsed),
		errors.Is(err, model.ErrAlreadyMember),
		errors.Is(err, model.ErrCannotChangeOwnRole),
		errors.Is(err, model.ErrInvalidInviteRole),
		errors.Is(err, model.ErrCannotRemoveSelf),
		errors.Is(err, model.ErrWorkspaceNameRequired),
		errors.Is(err, model.ErrWorkspaceNameTooLong),
		errors.Is(err, model.ErrWorkspaceDescriptionRequired),
		errors.Is(err, model.ErrWorkspaceDescriptionTooLong),
		errors.Is(err, model.ErrCategoryNameRequired),
		errors.Is(err, model.ErrCategoryNameTooLong),
		errors.Is(err, model.ErrCategoryBudgetsExceed),
		errors.Is(err, model.ErrCategoryUndeletable),
		errors.Is(err, model.ErrRuleInvalidPattern),
		errors.Is(err, model.ErrRulePatternTooLong),
		errors.Is(err, model.ErrTransactionAmountTooLarge),
		errors.Is(err, model.ErrTransactionDateTooFarInPast),
		errors.Is(err, model.ErrTransactionDateTooFarInFuture),
		errors.Is(err, model.ErrTransactionNoEntries),
		errors.Is(err, model.ErrEntrySumMismatch),
		errors.Is(err, model.ErrBankAccountAlreadyExists),
		errors.Is(err, model.ErrBankConnectionExpired),
		errors.Is(err, model.ErrUnsupportedLanguage):
		respondAppError(w, http.StatusBadRequest, err)
	case errors.Is(err, model.ErrDuplicateFingerprint):
		respondAppError(w, http.StatusConflict, err)
	default:
		slog.Error("unhandled service error", "error", err)
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "internal server error")
	}
}
