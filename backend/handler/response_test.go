package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
		wantCT     string
	}{
		{
			name:       "writes JSON with 200",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantCT:     "application/json",
		},
		{
			name:       "writes JSON with 201",
			status:     http.StatusCreated,
			data:       map[string]int{"id": 42},
			wantStatus: http.StatusCreated,
			wantCT:     "application/json",
		},
		{
			name:       "writes JSON array",
			status:     http.StatusOK,
			data:       []string{"a", "b", "c"},
			wantStatus: http.StatusOK,
			wantCT:     "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			ct := w.Header().Get("Content-Type")
			if ct != tt.wantCT {
				t.Errorf("want Content-Type %q, got %q", tt.wantCT, ct)
			}

			var decoded any
			if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
				t.Errorf("response body is not valid JSON: %v", err)
			}
		})
	}
}

func TestRespondJSON_EncodesDataCorrectly(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{
		"id":   123,
		"name": "test",
		"tags": []string{"a", "b"},
	}

	respondJSON(w, http.StatusOK, data)

	var decoded map[string]any
	if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if int(decoded["id"].(float64)) != 123 {
		t.Errorf("want id=123, got %v", decoded["id"])
	}
	if decoded["name"] != "test" {
		t.Errorf("want name=test, got %v", decoded["name"])
	}
}

func TestRespondCodedError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		code        string
		message     string
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "400 bad request",
			status:      http.StatusBadRequest,
			code:        "INVALID_REQUEST",
			message:     "invalid input",
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST",
			wantMessage: "invalid input",
		},
		{
			name:        "401 unauthorized",
			status:      http.StatusUnauthorized,
			code:        "UNAUTHORIZED",
			message:     "missing token",
			wantStatus:  http.StatusUnauthorized,
			wantCode:    "UNAUTHORIZED",
			wantMessage: "missing token",
		},
		{
			name:        "404 not found",
			status:      http.StatusNotFound,
			code:        "WORKSPACE_NOT_FOUND",
			message:     "workspace not found",
			wantStatus:  http.StatusNotFound,
			wantCode:    "WORKSPACE_NOT_FOUND",
			wantMessage: "workspace not found",
		},
		{
			name:        "500 internal error",
			status:      http.StatusInternalServerError,
			code:        "INTERNAL_ERROR",
			message:     "internal server error",
			wantStatus:  http.StatusInternalServerError,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondCodedError(w, tt.status, tt.code, tt.message)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			ct := w.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("want Content-Type application/json, got %q", ct)
			}

			var env codedErrorEnvelope
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatalf("decoding error envelope: %v", err)
			}

			if env.Code != tt.wantCode {
				t.Errorf("want error code %q, got %q", tt.wantCode, env.Code)
			}
			if env.Message != tt.wantMessage {
				t.Errorf("want error message %q, got %q", tt.wantMessage, env.Message)
			}
		})
	}
}
