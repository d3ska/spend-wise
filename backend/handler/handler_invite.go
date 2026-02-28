package handler

import (
	"errors"
	"net/http"

	"backend/auth"
	"backend/model"
	"backend/service"

	"github.com/go-chi/chi/v5"
)

// InviteHandler handles invite HTTP endpoints.
type InviteHandler struct {
	inviteSvc *service.InviteService
	wsSvc     *service.WorkspaceService
	hub       *SSEHub
}

// NewInviteHandler creates an InviteHandler.
func NewInviteHandler(inviteSvc *service.InviteService, wsSvc *service.WorkspaceService, hub *SSEHub) *InviteHandler {
	return &InviteHandler{inviteSvc: inviteSvc, wsSvc: wsSvc, hub: hub}
}

// CreateInvite handles POST /api/v1/workspaces/{id}/invites.
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
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
		Role           string `json:"role"`
		ExpiresInHours int    `json:"expires_in_hours"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	inv, err := h.inviteSvc.CreateInvite(r.Context(), userID, model.WorkspaceID(wsID), service.CreateInviteInput{
		Role:           model.MemberRole(body.Role),
		ExpiresInHours: body.ExpiresInHours,
	})
	if err != nil {
		handleInviteError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, inviteResponse(inv))
}

// PreviewInvite handles GET /api/v1/invites/{code} (public endpoint).
func (h *InviteHandler) PreviewInvite(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invite code is required")
		return
	}

	preview, err := h.inviteSvc.PreviewInvite(r.Context(), code)
	if err != nil {
		handleInviteError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"workspace_name": preview.WorkspaceName,
		"inviter_name":   preview.InviterName,
		"role":           preview.Role,
		"expires_at":     preview.ExpiresAt,
		"expired":        preview.Expired,
		"used":           preview.Used,
	})
}

// AcceptInvite handles POST /api/v1/invites/{code}/accept.
func (h *InviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	code := chi.URLParam(r, "code")
	if code == "" {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invite code is required")
		return
	}

	ws, err := h.inviteSvc.AcceptInvite(r.Context(), userID, code)
	if err != nil {
		handleInviteError(w, err)
		return
	}

	h.hub.Broadcast(int64(ws.ID), "member_changed")
	respondJSON(w, http.StatusOK, workspaceResponse(ws))
}

// ListMembers handles GET /api/v1/workspaces/{id}/members.
func (h *InviteHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.wsSvc.ListMembersWithProfiles(r.Context(), userID, model.WorkspaceID(wsID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(members))
	for _, m := range members {
		result = append(result, map[string]any{
			"user_id":      m.UserID,
			"display_name": m.DisplayName,
			"email":        m.Email,
			"avatar_url":   m.AvatarURL,
			"role":         m.Role,
			"created_at":   m.CreatedAt,
		})
	}
	respondJSON(w, http.StatusOK, result)
}

// UpdateMemberRole handles PUT /api/v1/workspaces/{id}/members/{userID}.
func (h *InviteHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Role string `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	if err := h.wsSvc.UpdateMemberRole(r.Context(), userID, model.WorkspaceID(wsID), model.UserID(targetUserID), model.MemberRole(body.Role)); err != nil {
		handleInviteError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "member_changed")
	w.WriteHeader(http.StatusNoContent)
}

func inviteResponse(inv model.Invite) map[string]any {
	return map[string]any{
		"id":           inv.ID,
		"workspace_id": inv.WorkspaceID,
		"code":         inv.Code,
		"role":         inv.Role,
		"expires_at":   inv.ExpiresAt,
		"created_at":   inv.CreatedAt,
	}
}

func handleInviteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrInviteNotFound):
		respondAppError(w, http.StatusNotFound, err)
	case errors.Is(err, model.ErrInviteExpired):
		respondAppError(w, http.StatusGone, err)
	case errors.Is(err, model.ErrInviteUsed):
		respondAppError(w, http.StatusGone, err)
	case errors.Is(err, model.ErrAlreadyMember):
		respondAppError(w, http.StatusConflict, err)
	case errors.Is(err, model.ErrCannotChangeOwnRole):
		respondAppError(w, http.StatusBadRequest, err)
	case errors.Is(err, model.ErrInvalidInviteRole):
		respondAppError(w, http.StatusBadRequest, err)
	default:
		handleServiceError(w, err)
	}
}
