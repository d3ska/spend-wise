package model

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors_Identity(t *testing.T) {
	sentinels := []error{
		ErrInvalidCredentials,
		ErrEmailAlreadyExists,
		ErrUserNotFound,
		ErrWorkspaceNotFound,
		ErrWorkspaceNameRequired,
		ErrWorkspaceDescriptionRequired,
		ErrNotWorkspaceMember,
		ErrInsufficientPermission,
		ErrTransactionNotFound,
		ErrTransactionNoEntries,
		ErrEntrySumMismatch,
		ErrDuplicateFingerprint,
		ErrCategoryNotFound,
		ErrCategoryUndeletable,
		ErrFundingNotFound,
		ErrCategoryBudgetsExceed,
		ErrRuleNotFound,
		ErrBankConnectionNotFound,
		ErrBankAccountNotFound,
		ErrBankAccountAlreadyExists,
		ErrBankConnectionExpired,
	}

	t.Run("direct identity", func(t *testing.T) {
		for _, err := range sentinels {
			if !errors.Is(err, err) {
				t.Errorf("errors.Is(%v, %v) got false, want true", err, err)
			}
		}
	})

	t.Run("wrapped identity", func(t *testing.T) {
		for _, err := range sentinels {
			wrapped := fmt.Errorf("context: %w", err)
			if !errors.Is(wrapped, err) {
				t.Errorf("errors.Is(wrapped %v, %v) got false, want true", err, err)
			}
		}
	})

	t.Run("distinct errors are not equal", func(t *testing.T) {
		for i, a := range sentinels {
			for j, b := range sentinels {
				if i != j && errors.Is(a, b) {
					t.Errorf("errors.Is(%v, %v) got true, want false", a, b)
				}
			}
		}
	})
}

func TestAppError_ErrorReturnsMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{"workspace not found", ErrWorkspaceNotFound, "workspace not found"},
		{"insufficient permission", ErrInsufficientPermission, "insufficient permission"},
		{"unsupported language", ErrUnsupportedLanguage, "unsupported language"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppError_CodeAccessible(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		code string
	}{
		{"workspace not found", ErrWorkspaceNotFound, "WORKSPACE_NOT_FOUND"},
		{"insufficient permission", ErrInsufficientPermission, "INSUFFICIENT_PERMISSION"},
		{"invalid credentials", ErrInvalidCredentials, "INVALID_CREDENTIALS"},
		{"transaction not found", ErrTransactionNotFound, "TRANSACTION_NOT_FOUND"},
		{"unsupported language", ErrUnsupportedLanguage, "UNSUPPORTED_LANGUAGE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Code; got != tt.code {
				t.Errorf("AppError.Code = %q, want %q", got, tt.code)
			}
		})
	}
}

func TestAppError_IsWithWrappedError(t *testing.T) {
	wrapped := fmt.Errorf("service: %w", ErrWorkspaceNotFound)

	var appErr *AppError
	if !errors.As(wrapped, &appErr) {
		t.Fatal("errors.As failed to extract AppError from wrapped error")
	}
	if appErr.Code != "WORKSPACE_NOT_FOUND" {
		t.Errorf("extracted code = %q, want %q", appErr.Code, "WORKSPACE_NOT_FOUND")
	}
}
