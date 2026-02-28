package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backend/auth"
	"backend/model"
	"backend/service"

	"github.com/shopspring/decimal"
)

// TransactionHandler handles transaction HTTP endpoints.
type TransactionHandler struct {
	svc *service.TransactionService
	hub *SSEHub
}

// NewTransactionHandler creates a TransactionHandler.
func NewTransactionHandler(svc *service.TransactionService, hub *SSEHub) *TransactionHandler {
	return &TransactionHandler{svc: svc, hub: hub}
}

// CreateTransaction handles POST /api/v1/workspaces/{id}/transactions.
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
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

	var body createTransactionRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	input, err := body.toInput()
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, err.Error())
		return
	}

	tx, err := h.svc.Create(r.Context(), userID, model.WorkspaceID(wsID), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	respondJSON(w, http.StatusCreated, transactionResponse(tx))
}

// ListTransactions handles GET /api/v1/workspaces/{id}/transactions.
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
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

	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid 'from' date, expected YYYY-MM-DD")
		return
	}
	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid 'to' date, expected YYYY-MM-DD")
		return
	}
	// The SQL queries use half-open intervals (date < $to), so add 1 day
	// to include the end date the caller specified.
	to = to.AddDate(0, 0, 1)

	limit := int32(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.ParseInt(l, 10, 32)
		if err != nil || parsed < 1 || parsed > 200 {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "limit must be between 1 and 200")
			return
		}
		limit = int32(parsed)
	}

	offset := int32(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		parsed, err := strconv.ParseInt(o, 10, 32)
		if err != nil || parsed < 0 {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "offset must be non-negative")
			return
		}
		offset = int32(parsed)
	}

	var txType *model.TransactionType
	if t := r.URL.Query().Get("type"); t != "" {
		tt := model.TransactionType(t)
		if tt != model.TypeExpense && tt != model.TypeIncome && tt != model.TypeTransfer {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "type must be 'expense', 'income', or 'transfer'")
			return
		}
		txType = &tt
	}

	var currency *string
	if c := r.URL.Query().Get("currency"); c != "" {
		currency = &c
	}

	var categoryID *model.CategoryID
	if c := r.URL.Query().Get("category_id"); c != "" {
		parsed, err := strconv.ParseInt(c, 10, 64)
		if err != nil || parsed < 1 {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid category_id")
			return
		}
		cid := model.CategoryID(parsed)
		categoryID = &cid
	}

	var bankAccountID *model.BankAccountID
	if b := r.URL.Query().Get("bank_account_id"); b != "" {
		parsed, err := strconv.ParseInt(b, 10, 64)
		if err != nil || parsed < 1 {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid bank_account_id")
			return
		}
		bid := model.BankAccountID(parsed)
		bankAccountID = &bid
	}

	var amountMin *decimal.Decimal
	if v := r.URL.Query().Get("amount_min"); v != "" {
		d, err := decimal.NewFromString(v)
		if err != nil || d.IsNegative() {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "amount_min must be a non-negative number")
			return
		}
		amountMin = &d
	}

	var amountMax *decimal.Decimal
	if v := r.URL.Query().Get("amount_max"); v != "" {
		d, err := decimal.NewFromString(v)
		if err != nil || d.IsNegative() {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "amount_max must be a non-negative number")
			return
		}
		amountMax = &d
	}

	sortBy := "date"
	if s := r.URL.Query().Get("sort_by"); s != "" {
		switch s {
		case "date", "amount", "description":
			sortBy = s
		default:
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "sort_by must be 'date', 'amount', or 'description'")
			return
		}
	}

	sortOrder := "desc"
	if s := r.URL.Query().Get("sort_order"); s != "" {
		switch s {
		case "asc", "desc":
			sortOrder = s
		default:
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "sort_order must be 'asc' or 'desc'")
			return
		}
	}

	txs, err := h.svc.ListTransactions(r.Context(), userID, model.WorkspaceID(wsID), service.ListTransactionsInput{
		From:          from,
		To:            to,
		Limit:         limit,
		Offset:        offset,
		Type:          txType,
		Currency:      currency,
		CategoryID:    categoryID,
		BankAccountID: bankAccountID,
		AmountMin:     amountMin,
		AmountMax:     amountMax,
		SortBy:        sortBy,
		SortOrder:     sortOrder,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(txs))
	for _, tx := range txs {
		result = append(result, transactionResponse(tx))
	}
	respondJSON(w, http.StatusOK, result)
}

// GetTransaction handles GET /api/v1/workspaces/{id}/transactions/{txID}.
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
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

	txID, err := parseIDParam(r, "txID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid transaction id")
		return
	}

	tx, err := h.svc.GetTransaction(r.Context(), userID, model.WorkspaceID(wsID), model.TransactionID(txID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, transactionResponse(tx))
}

// DeleteTransaction handles DELETE /api/v1/workspaces/{id}/transactions/{txID}.
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
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

	txID, err := parseIDParam(r, "txID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid transaction id")
		return
	}

	if err := h.svc.DeleteTransaction(r.Context(), userID, model.WorkspaceID(wsID), model.TransactionID(txID)); err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	w.WriteHeader(http.StatusNoContent)
}

// UpdateTransaction handles PUT /api/v1/workspaces/{id}/transactions/{txID}.
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
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

	txID, err := parseIDParam(r, "txID")
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid transaction id")
		return
	}

	var body createTransactionRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	input, err := body.toInput()
	if err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, err.Error())
		return
	}

	tx, err := h.svc.UpdateTransaction(r.Context(), userID, model.WorkspaceID(wsID), model.TransactionID(txID), service.UpdateTransactionInput{
		Description: input.Description,
		Date:        input.Date,
		Type:        input.Type,
		Notes:       input.Notes,
		Entries:     input.Entries,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	respondJSON(w, http.StatusOK, transactionResponse(tx))
}

// maxImportBatchSize is the maximum number of transactions allowed in a single import request.
const maxImportBatchSize = 500

// ImportTransactions handles POST /api/v1/workspaces/{id}/transactions/import.
func (h *TransactionHandler) ImportTransactions(w http.ResponseWriter, r *http.Request) {
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

	var body []importTransactionRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	if len(body) > maxImportBatchSize {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, fmt.Sprintf("import batch too large: maximum %d transactions per request", maxImportBatchSize))
		return
	}

	inputs := make([]service.ImportTransactionInput, 0, len(body))
	for _, item := range body {
		input, err := item.toInput()
		if err != nil {
			respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, err.Error())
			return
		}
		inputs = append(inputs, input)
	}

	result, err := h.svc.ImportTransactions(r.Context(), userID, model.WorkspaceID(wsID), inputs)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	txResponses := make([]map[string]any, 0, len(result.Transactions))
	for _, tx := range result.Transactions {
		txResponses = append(txResponses, transactionResponse(tx))
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	respondJSON(w, http.StatusOK, map[string]any{
		"imported":     result.Imported,
		"skipped":      result.Skipped,
		"transactions": txResponses,
	})
}

// ── Request types ──

type moneyRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type entryRequest struct {
	CategoryID    *int64       `json:"category_id"`
	ParticipantID *int64       `json:"participant_id"`
	Amount        moneyRequest `json:"amount"`
	Note          string       `json:"note"`
}

type createTransactionRequest struct {
	Description string         `json:"description"`
	Date        string         `json:"date"`
	TotalAmount moneyRequest   `json:"total_amount"`
	Type        string         `json:"type"`
	Notes       string         `json:"notes"`
	Entries     []entryRequest `json:"entries"`
}

type bulkCategorizeRequest struct {
	IDs        []int64 `json:"ids"`
	CategoryID int64   `json:"category_id"`
}

type bulkDeleteRequest struct {
	IDs []int64 `json:"ids"`
}

type importTransactionRequest struct {
	createTransactionRequest
	Fingerprint string `json:"fingerprint"`
}

func (r createTransactionRequest) toInput() (service.CreateTransactionInput, error) {
	date, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return service.CreateTransactionInput{}, fmt.Errorf("invalid date: %s", r.Date)
	}

	totalAmount, err := parseMoney(r.TotalAmount)
	if err != nil {
		return service.CreateTransactionInput{}, err
	}

	entries := make([]service.CreateEntryInput, 0, len(r.Entries))
	for _, e := range r.Entries {
		entryAmount, err := parseMoney(e.Amount)
		if err != nil {
			return service.CreateTransactionInput{}, err
		}
		var participantID *model.UserID
		if e.ParticipantID != nil {
			uid := model.UserID(*e.ParticipantID)
			participantID = &uid
		}
		var categoryID model.CategoryID
		if e.CategoryID != nil {
			categoryID = model.CategoryID(*e.CategoryID)
		}
		entries = append(entries, service.CreateEntryInput{
			CategoryID:    categoryID,
			ParticipantID: participantID,
			Amount:        entryAmount,
			Note:          e.Note,
		})
	}

	txType := model.TransactionType(r.Type)
	if txType != "" && txType != model.TypeExpense && txType != model.TypeIncome && txType != model.TypeTransfer {
		return service.CreateTransactionInput{}, fmt.Errorf("invalid type: %s", r.Type)
	}

	return service.CreateTransactionInput{
		Description: r.Description,
		Date:        date,
		TotalAmount: totalAmount,
		Type:        txType,
		Notes:       r.Notes,
		Entries:     entries,
	}, nil
}

func (r importTransactionRequest) toInput() (service.ImportTransactionInput, error) {
	base, err := r.createTransactionRequest.toInput()
	if err != nil {
		return service.ImportTransactionInput{}, err
	}
	return service.ImportTransactionInput{
		CreateTransactionInput: base,
		Fingerprint:            r.Fingerprint,
	}, nil
}

// BulkCategorizeTransactions handles POST /api/v1/workspaces/{id}/transactions/bulk-categorize.
func (h *TransactionHandler) BulkCategorizeTransactions(w http.ResponseWriter, r *http.Request) {
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

	var body bulkCategorizeRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	if len(body.IDs) == 0 {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "ids must not be empty")
		return
	}
	if body.CategoryID <= 0 {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "category_id must be positive")
		return
	}

	txIDs := make([]model.TransactionID, len(body.IDs))
	for i, id := range body.IDs {
		txIDs[i] = model.TransactionID(id)
	}

	updated, err := h.svc.BulkCategorizeTransactions(r.Context(), userID, model.WorkspaceID(wsID), txIDs, model.CategoryID(body.CategoryID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	respondJSON(w, http.StatusOK, map[string]any{"updated": updated})
}

// BulkDeleteTransactions handles POST /api/v1/workspaces/{id}/transactions/bulk-delete.
func (h *TransactionHandler) BulkDeleteTransactions(w http.ResponseWriter, r *http.Request) {
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

	var body bulkDeleteRequest
	if err := decodeJSON(r, &body); err != nil {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "invalid request body")
		return
	}

	if len(body.IDs) == 0 {
		respondCodedError(w, http.StatusBadRequest, model.CodeInvalidRequest, "ids must not be empty")
		return
	}

	txIDs := make([]model.TransactionID, len(body.IDs))
	for i, id := range body.IDs {
		txIDs[i] = model.TransactionID(id)
	}

	deleted, err := h.svc.BulkDeleteTransactions(r.Context(), userID, model.WorkspaceID(wsID), txIDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	respondJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

// ApplyRules handles POST /api/v1/workspaces/{id}/transactions/apply-rules.
func (h *TransactionHandler) ApplyRules(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.svc.ApplyRules(r.Context(), callerID, model.WorkspaceID(wsID))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.hub.Broadcast(wsID, "transaction_changed")
	respondJSON(w, http.StatusOK, result)
}

func parseMoney(m moneyRequest) (model.Money, error) {
	d, err := decimal.NewFromString(m.Amount)
	if err != nil {
		return model.Money{}, fmt.Errorf("invalid amount: %s", m.Amount)
	}
	currency := m.Currency
	if currency == "" {
		currency = model.DefaultCurrency
	}
	return model.NewMoney(d, currency), nil
}

// ── Response helpers ──

func transactionResponse(tx model.Transaction) map[string]any {
	entries := make([]map[string]any, 0, len(tx.Entries))
	for _, e := range tx.Entries {
		entry := map[string]any{
			"id":          e.ID,
			"category_id": e.CategoryID,
			"amount": map[string]any{
				"amount":   e.Amount.Amount().StringFixed(2),
				"currency": e.Amount.Currency(),
			},
			"note": e.Note,
		}
		if e.ParticipantID != nil {
			entry["participant_id"] = *e.ParticipantID
		}
		entries = append(entries, entry)
	}

	resp := map[string]any{
		"id":           tx.ID,
		"workspace_id": tx.WorkspaceID,
		"created_by":   tx.CreatedBy,
		"description":  tx.Description,
		"date":         tx.Date.Format("2006-01-02"),
		"total_amount": map[string]any{
			"amount":   tx.TotalAmount.Amount().StringFixed(2),
			"currency": tx.TotalAmount.Currency(),
		},
		"source":            tx.Source,
		"type":              tx.Type,
		"notes":             tx.Notes,
		"entries":           entries,
		"bank_name":         tx.BankName,
		"iban":              tx.IBAN,
		"counterparty_iban": tx.CounterpartyIBAN,
		"bank_account_id":   tx.BankAccountID,
		"created_at":        tx.CreatedAt,
		"updated_at":        tx.UpdatedAt,
	}
	return resp
}
