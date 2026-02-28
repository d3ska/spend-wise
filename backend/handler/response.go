package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"backend/model"
)

type codedErrorEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondJSON writes a JSON response with the given status code and data.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("writing JSON response", "error", err)
	}
}

// respondCodedError writes an error response with a machine-readable code and message.
func respondCodedError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, codedErrorEnvelope{Code: code, Message: message})
}

// respondAppError writes an error response extracting the code from an AppError.
// Falls back to INTERNAL_ERROR for non-AppError values.
func respondAppError(w http.ResponseWriter, status int, err error) {
	if appErr, ok := err.(*model.AppError); ok {
		respondCodedError(w, status, appErr.Code, appErr.Message)
		return
	}
	respondCodedError(w, status, model.CodeInternalError, "internal server error")
}
