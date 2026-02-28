package handler

import (
	"net/http"

	"backend/auth"
	"backend/model"
	"backend/service"
)

// CategoryHandler handles category HTTP endpoints.
type CategoryHandler struct {
	svc *service.WorkspaceService
	hub *SSEHub
}

// NewCategoryHandler creates a CategoryHandler.
func NewCategoryHandler(svc *service.WorkspaceService, hub *SSEHub) *CategoryHandler {
	return &CategoryHandler{svc: svc, hub: hub}
}

// CreateCategory handles POST /api/v1/workspaces/{id}/categories.
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
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
		Name string `json:"name"`
		Icon string `json:"icon"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), userID, model.WorkspaceID(wsID), service.CreateCategoryInput{
		Name: body.Name,
		Icon: body.Icon,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "category_changed")
	respondJSON(w, http.StatusCreated, categoryResponse(cat))
}

// ListCategories handles GET /api/v1/workspaces/{id}/categories.
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
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

	categories, err := h.svc.ListCategories(r.Context(), userID, model.WorkspaceID(wsID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(categories))
	for _, cat := range categories {
		result = append(result, categoryResponse(cat))
	}
	respondJSON(w, http.StatusOK, result)
}

// UpdateCategory handles PUT /api/v1/workspaces/{id}/categories/{catID}.
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
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

	catID, err := parseIDParam(r, "catID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid category id")
		return
	}

	var body struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	cat, err := h.svc.UpdateCategory(r.Context(), userID, model.WorkspaceID(wsID), model.CategoryID(catID), service.UpdateCategoryInput{
		Name: body.Name,
		Icon: body.Icon,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "category_changed")
	respondJSON(w, http.StatusOK, categoryResponse(cat))
}

// DeleteCategory handles DELETE /api/v1/workspaces/{id}/categories/{catID}.
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
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

	catID, err := parseIDParam(r, "catID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid category id")
		return
	}

	if err := h.svc.DeleteCategory(r.Context(), userID, model.WorkspaceID(wsID), model.CategoryID(catID)); err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "category_changed")
	w.WriteHeader(http.StatusNoContent)
}

func categoryResponse(cat model.Category) map[string]any {
	return map[string]any{
		"id":           cat.ID,
		"workspace_id": cat.WorkspaceID,
		"name":         cat.Name,
		"icon":         cat.Icon,
		"slug":         cat.Slug,
		"created_at":   cat.CreatedAt,
		"updated_at":   cat.UpdatedAt,
	}
}
