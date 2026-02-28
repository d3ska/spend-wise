package handler

import (
	"fmt"
	"net/http"
	"time"

	"backend/auth"
	"backend/model"
	"backend/service"
)

// SSEHandler handles the Server-Sent Events endpoint.
type SSEHandler struct {
	hub *SSEHub
	svc *service.WorkspaceService
}

// NewSSEHandler creates an SSEHandler.
func NewSSEHandler(hub *SSEHub, svc *service.WorkspaceService) *SSEHandler {
	return &SSEHandler{hub: hub, svc: svc}
}

// StreamEvents handles GET /api/v1/workspaces/{id}/events.
func (h *SSEHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
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

	// Verify caller is a workspace member.
	if _, err := h.svc.GetWorkspace(r.Context(), userID, model.WorkspaceID(wsID)); err != nil {
		handleServiceError(w, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "streaming not supported")
		return
	}

	// Disable write deadline for this long-lived connection.
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch, unsubscribe, err := h.hub.Subscribe(wsID)
	if err != nil {
		respondCodedError(w, http.StatusTooManyRequests, model.CodeInvalidRequest, "too many SSE connections for this workspace")
		return
	}
	defer unsubscribe()

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case eventType, ok := <-ch:
			if !ok {
				return // hub closed, shutdown in progress
			}
			fmt.Fprintf(w, "event: %s\ndata: {}\n\n", eventType)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprint(w, ":heartbeat\n\n")
			flusher.Flush()
		}
	}
}
