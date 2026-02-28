package handler

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"time"

	"backend/auth"
	"backend/model"
	"backend/service"
)

// BankHandler handles bank connection and account linking HTTP endpoints.
type BankHandler struct {
	connSvc    *service.BankConnectionService
	linkingSvc *service.BankAccountLinkingService
	syncSvc    *service.BankSyncService
}

// NewBankHandler creates a BankHandler.
func NewBankHandler(
	connSvc *service.BankConnectionService,
	linkingSvc *service.BankAccountLinkingService,
	syncSvc *service.BankSyncService,
) *BankHandler {
	return &BankHandler{
		connSvc:    connSvc,
		linkingSvc: linkingSvc,
		syncSvc:    syncSvc,
	}
}

type institutionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
	BIC  string `json:"bic"`
}

// ListInstitutions handles GET /api/v1/banks?country=PL.
func (h *BankHandler) ListInstitutions(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		country = "PL"
	}

	aspsps, err := h.connSvc.ListInstitutions(r.Context(), country)
	if err != nil {
		slog.Error("failed to fetch institutions", "error", err, "country", country)
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "failed to fetch institutions")
		return
	}

	result := make([]institutionResponse, 0, len(aspsps))
	for _, a := range aspsps {
		slog.Debug("ASPSP found", "name", a.Name, "country", a.Country)
		result = append(result, institutionResponse{
			ID:   a.Name,
			Name: a.Name,
			Logo: a.Logo,
			BIC:  a.BIC,
		})
	}

	respondJSON(w, http.StatusOK, result)
}

type initiateConnectionRequest struct {
	InstitutionID   string `json:"institution_id"`
	InstitutionName string `json:"institution_name"`
	RedirectURL     string `json:"redirect_url"`
	Country         string `json:"country"`
}

type initiateConnectionResponse struct {
	ConnectionID int64  `json:"connection_id"`
	AuthLink     string `json:"auth_link"`
}

// InitiateConnection handles POST /api/v1/bank-connections.
func (h *BankHandler) InitiateConnection(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	var body initiateConnectionRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	result, err := h.connSvc.Initiate(r.Context(), model.UserID(callerID), body.InstitutionID, body.InstitutionName, body.RedirectURL, body.Country)
	if err != nil {
		slog.Error("failed to initiate bank connection", "error", err, "institution_id", body.InstitutionID)
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, initiateConnectionResponse{
		ConnectionID: int64(result.Connection.ID),
		AuthLink:     result.AuthLink,
	})
}

type completeConnectionRequest struct {
	Code string `json:"code"`
}

// CompleteConnection handles POST /api/v1/bank-connections/{id}/complete.
func (h *BankHandler) CompleteConnection(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid connection id")
		return
	}

	// Accept code from request body or query param.
	code := r.URL.Query().Get("code")
	if code == "" {
		var body completeConnectionRequest
		if err := decodeJSON(r, &body); err == nil {
			code = body.Code
		}
	}
	if code == "" {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "missing authorization code")
		return
	}

	conn, err := h.connSvc.GetConnection(r.Context(), model.UserID(callerID), model.BankConnectionID(connID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	accounts, err := h.connSvc.Complete(r.Context(), model.UserID(callerID), model.BankConnectionID(connID), code)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, toBankAccountResponses(accounts, conn.InstitutionName))
}

// CompleteConnectionByCode handles POST /api/v1/bank-connections/complete.
// Accepts just a code and finds the matching connection for the user.
func (h *BankHandler) CompleteConnectionByCode(w http.ResponseWriter, r *http.Request) {
	slog.Info("CompleteConnectionByCode: handler entered")

	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		slog.Warn("CompleteConnectionByCode: no auth context")
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}
	slog.Info("CompleteConnectionByCode: authenticated", "user_id", callerID)

	var body completeConnectionRequest
	if err := decodeJSON(r, &body); err != nil || body.Code == "" {
		slog.Warn("CompleteConnectionByCode: missing or invalid code", "decode_err", err)
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "missing authorization code")
		return
	}
	slog.Info("CompleteConnectionByCode: received code", "code_len", len(body.Code))

	accounts, institutionName, err := h.connSvc.CompleteByCode(r.Context(), model.UserID(callerID), body.Code)
	if err != nil {
		slog.Error("CompleteConnectionByCode: service error", "error", err)
		handleServiceError(w, err)
		return
	}

	slog.Info("CompleteConnectionByCode: success", "accounts_count", len(accounts))
	respondJSON(w, http.StatusOK, toBankAccountResponses(accounts, institutionName))
}

type bankConnectionResponse struct {
	ID              int64  `json:"id"`
	InstitutionID   string `json:"institution_id"`
	InstitutionName string `json:"institution_name"`
	Status          string `json:"status"`
	AuthExpiresAt   string `json:"auth_expires_at"`
	DaysRemaining   int    `json:"days_remaining"`
	ExpiryStatus    string `json:"expiry_status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toBankConnectionResponse(bc model.BankConnection) bankConnectionResponse {
	daysRemaining := int(math.Ceil(time.Until(bc.AuthExpiresAt).Hours() / 24))
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	var expiryStatus string
	switch {
	case daysRemaining <= 0:
		expiryStatus = "expired"
	case daysRemaining <= 14:
		expiryStatus = "warning"
	default:
		expiryStatus = "ok"
	}

	return bankConnectionResponse{
		ID:              int64(bc.ID),
		InstitutionID:   bc.InstitutionID,
		InstitutionName: bc.InstitutionName,
		Status:          string(bc.Status),
		AuthExpiresAt:   bc.AuthExpiresAt.Format(time.RFC3339),
		DaysRemaining:   daysRemaining,
		ExpiryStatus:    expiryStatus,
		CreatedAt:       bc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       bc.UpdatedAt.Format(time.RFC3339),
	}
}

// ListConnections handles GET /api/v1/bank-connections.
func (h *BankHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connections, err := h.connSvc.ListByUser(r.Context(), model.UserID(callerID))
	if err != nil {
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "failed to list connections")
		return
	}

	result := make([]bankConnectionResponse, 0, len(connections))
	for _, conn := range connections {
		result = append(result, toBankConnectionResponse(conn))
	}

	respondJSON(w, http.StatusOK, result)
}

// DeleteConnection handles DELETE /api/v1/bank-connections/{id}.
func (h *BankHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid connection id")
		return
	}

	if err := h.connSvc.Delete(r.Context(), model.UserID(callerID), model.BankConnectionID(connID)); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeactivateConnection handles POST /api/v1/bank-connections/{id}/deactivate.
func (h *BankHandler) DeactivateConnection(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid connection id")
		return
	}

	if err := h.connSvc.Deactivate(r.Context(), model.UserID(callerID), model.BankConnectionID(connID)); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReactivateConnection handles POST /api/v1/bank-connections/{id}/reactivate.
func (h *BankHandler) ReactivateConnection(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid connection id")
		return
	}

	if err := h.connSvc.Reactivate(r.Context(), model.UserID(callerID), model.BankConnectionID(connID)); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type reconnectRequest struct {
	RedirectURL string `json:"redirect_url"`
}

type reconnectResponse struct {
	AuthLink string `json:"auth_link"`
}

// ReconnectConnection handles POST /api/v1/bank-connections/{id}/reconnect.
func (h *BankHandler) ReconnectConnection(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid connection id")
		return
	}

	var body reconnectRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	authLink, err := h.connSvc.Reconnect(r.Context(), model.UserID(callerID), model.BankConnectionID(connID), body.RedirectURL)
	if err != nil {
		if errors.Is(err, model.ErrBankConnectionExpired) {
			respondAppError(w, http.StatusConflict, err)
			return
		}
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, reconnectResponse{AuthLink: authLink})
}

// TriggerSync handles POST /api/v1/bank-connections/{id}/sync.
func (h *BankHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	connID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid connection id")
		return
	}

	// Verify ownership.
	connections, err := h.connSvc.ListByUser(r.Context(), model.UserID(callerID))
	if err != nil {
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "failed to list connections")
		return
	}

	var target *model.BankConnection
	for _, c := range connections {
		if c.ID == model.BankConnectionID(connID) {
			target = &c
			break
		}
	}
	if target == nil {
		respondCodedError(w, http.StatusNotFound, model.ErrBankConnectionNotFound.Code, "bank connection not found")
		return
	}
	if target.Status == model.BankConnectionExpired {
		respondCodedError(w, http.StatusConflict, model.ErrBankConnectionExpired.Code, "bank connection has expired, please reconnect")
		return
	}

	// Trigger sync for this connection only, asynchronously with a detached
	// context so it survives after the HTTP response is sent.
	go h.syncSvc.SyncConnection(context.Background(), *target)

	w.WriteHeader(http.StatusAccepted)
}

// ListUserBankAccounts handles GET /api/v1/bank-accounts.
func (h *BankHandler) ListUserBankAccounts(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	accounts, err := h.linkingSvc.ListByUser(r.Context(), model.UserID(callerID))
	if err != nil {
		respondCodedError(w, http.StatusInternalServerError, model.CodeInternalError, "failed to list bank accounts")
		return
	}

	result := make([]linkedBankAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		var lastSynced *string
		if a.LastSyncedAt != nil {
			s := a.LastSyncedAt.Format(time.RFC3339)
			lastSynced = &s
		}
		result = append(result, linkedBankAccountResponse{
			ID:               int64(a.ID),
			BankConnectionID: int64(a.BankConnectionID),
			InstitutionName:  a.InstitutionName,
			ExternalID:       a.ExternalID,
			IBAN:             a.IBAN,
			Currency:         a.Currency,
			Name:             a.Name,
			CustomName:       a.CustomName,
			DisplayName:      a.DisplayName(),
			LastSyncedAt:     lastSynced,
		})
	}

	respondJSON(w, http.StatusOK, result)
}

func toBankAccountResponses(accounts []model.BankAccount, institutionName string) []linkedBankAccountResponse {
	result := make([]linkedBankAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		var lastSynced *string
		if a.LastSyncedAt != nil {
			s := a.LastSyncedAt.Format(time.RFC3339)
			lastSynced = &s
		}
		result = append(result, linkedBankAccountResponse{
			ID:               int64(a.ID),
			BankConnectionID: int64(a.BankConnectionID),
			InstitutionName:  institutionName,
			ExternalID:       a.ExternalID,
			IBAN:             a.IBAN,
			Currency:         a.Currency,
			Name:             a.Name,
			CustomName:       a.CustomName,
			DisplayName:      a.DisplayName(),
			LastSyncedAt:     lastSynced,
		})
	}
	return result
}

type updateBankAccountCustomNameRequest struct {
	CustomName string `json:"custom_name"`
}

// UpdateBankAccountCustomName handles PATCH /api/v1/bank-accounts/{id}.
func (h *BankHandler) UpdateBankAccountCustomName(w http.ResponseWriter, r *http.Request) {
	callerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondCodedError(w, http.StatusUnauthorized, model.CodeUnauthorized, "missing or invalid token")
		return
	}

	baID, err := parseIDParam(r, "id")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid bank account id")
		return
	}

	var body updateBankAccountCustomNameRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	acct, err := h.linkingSvc.UpdateCustomName(r.Context(), model.UserID(callerID), model.BankAccountID(baID), body.CustomName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":           acct.ID,
		"name":         acct.Name,
		"custom_name":  acct.CustomName,
		"display_name": acct.DisplayName(),
	})
}

type linkBankAccountRequest struct {
	BankAccountID int64 `json:"bank_account_id"`
}

// LinkBankAccount handles POST /api/v1/workspaces/{id}/bank-accounts.
func (h *BankHandler) LinkBankAccount(w http.ResponseWriter, r *http.Request) {
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

	var body linkBankAccountRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	if err := h.linkingSvc.LinkToWorkspace(r.Context(), model.UserID(callerID), model.WorkspaceID(wsID), model.BankAccountID(body.BankAccountID)); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UnlinkBankAccount handles DELETE /api/v1/workspaces/{id}/bank-accounts/{bankAccountId}.
func (h *BankHandler) UnlinkBankAccount(w http.ResponseWriter, r *http.Request) {
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

	baID, err := parseIDParam(r, "bankAccountId")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid bank account id")
		return
	}

	if err := h.linkingSvc.UnlinkFromWorkspace(r.Context(), model.UserID(callerID), model.WorkspaceID(wsID), model.BankAccountID(baID)); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type linkedBankAccountResponse struct {
	ID               int64   `json:"id"`
	BankConnectionID int64   `json:"bank_connection_id"`
	InstitutionName  string  `json:"institution_name"`
	ExternalID       string  `json:"external_id"`
	IBAN             *string `json:"iban"`
	Currency         string  `json:"currency"`
	Name             string  `json:"name"`
	CustomName       string  `json:"custom_name"`
	DisplayName      string  `json:"display_name"`
	LastSyncedAt     *string `json:"last_synced_at"`
}

// ListLinkedBankAccounts handles GET /api/v1/workspaces/{id}/bank-accounts.
func (h *BankHandler) ListLinkedBankAccounts(w http.ResponseWriter, r *http.Request) {
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

	accounts, err := h.linkingSvc.ListByWorkspace(r.Context(), model.UserID(callerID), model.WorkspaceID(wsID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]linkedBankAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		var lastSynced *string
		if a.LastSyncedAt != nil {
			s := a.LastSyncedAt.Format(time.RFC3339)
			lastSynced = &s
		}
		result = append(result, linkedBankAccountResponse{
			ID:               int64(a.ID),
			BankConnectionID: int64(a.BankConnectionID),
			InstitutionName:  a.InstitutionName,
			ExternalID:       a.ExternalID,
			IBAN:             a.IBAN,
			Currency:         a.Currency,
			Name:             a.Name,
			CustomName:       a.CustomName,
			DisplayName:      a.DisplayName(),
			LastSyncedAt:     lastSynced,
		})
	}

	respondJSON(w, http.StatusOK, result)
}
