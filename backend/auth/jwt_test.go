package auth

import (
	"testing"
	"time"

	"backend/model"
)

func TestJWT_SignAndVerify(t *testing.T) {
	tests := []struct {
		name    string
		userID  model.UserID
		secret  string
		expiry  time.Duration
		wantErr bool
	}{
		{"valid round-trip", 42, "test-secret", 24 * time.Hour, false},
		{"different user ID", 999, "test-secret", 1 * time.Hour, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := SignToken(tt.userID, tt.secret, tt.expiry)
			if err != nil {
				t.Fatalf("SignToken: %v", err)
			}

			got, err := VerifyToken(token, tt.secret)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("VerifyToken: %v", err)
			}
			if got != tt.userID {
				t.Errorf("got UserID %d, want %d", got, tt.userID)
			}
		})
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	token, err := SignToken(1, "secret", -1*time.Hour)
	if err != nil {
		t.Fatalf("SignToken: %v", err)
	}
	_, err = VerifyToken(token, "secret")
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	token, err := SignToken(1, "secret-A", 1*time.Hour)
	if err != nil {
		t.Fatalf("SignToken: %v", err)
	}
	_, err = VerifyToken(token, "secret-B")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestJWT_MalformedToken(t *testing.T) {
	_, err := VerifyToken("not.a.jwt", "secret")
	if err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}
