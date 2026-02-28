package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		target    any
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid JSON body",
			body: `{"name":"test","value":42}`,
			target: &struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{},
			wantErr: false,
		},
		{
			name:      "invalid JSON",
			body:      `{"name":"test",}`,
			target:    &struct{}{},
			wantErr:   true,
			errSubstr: "decoding JSON",
		},
		{
			name: "unknown fields rejected",
			body: `{"name":"test","unknown":"field"}`,
			target: &struct {
				Name string `json:"name"`
			}{},
			wantErr:   true,
			errSubstr: "unknown field",
		},
		{
			name:      "body exceeding 1MB",
			body:      `{"data":"` + strings.Repeat("a", 1<<20) + `"}`,
			target:    &struct{}{},
			wantErr:   true,
			errSubstr: "decoding JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.body))
			err := decodeJSON(r, tt.target)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error containing %q, got nil", tt.errSubstr)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("want error containing %q, got %q", tt.errSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("want nil error, got %v", err)
				}
			}
		})
	}
}

func TestDecodeJSON_PopulatesStruct(t *testing.T) {
	body := `{"name":"workspace","description":"test workspace","count":5}`
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))

	var target struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Count       int    `json:"count"`
	}

	err := decodeJSON(r, &target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Name != "workspace" {
		t.Errorf("want Name=%q, got %q", "workspace", target.Name)
	}
	if target.Description != "test workspace" {
		t.Errorf("want Description=%q, got %q", "test workspace", target.Description)
	}
	if target.Count != 5 {
		t.Errorf("want Count=%d, got %d", 5, target.Count)
	}
}
