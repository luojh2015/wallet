package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(1001, "test message", "test detail")
	if err.Code != 1001 {
		t.Errorf("expected code 1001, got %d", err.Code)
	}
	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", err.Message)
	}
	if err.Detail != "test detail" {
		t.Errorf("expected detail 'test detail', got %q", err.Detail)
	}
}

func TestAppError_Error_WithDetail(t *testing.T) {
	err := NewAppError(2001, "钱包不存在", "ID=123")
	expected := "[2001] 钱包不存在: ID=123"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAppError_Error_WithoutDetail(t *testing.T) {
	err := NewAppError(2001, "钱包不存在", "")
	expected := "[2001] 钱包不存在"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"AppError", NewAppError(1001, "test", ""), true},
		{"standard error", errors.New("standard"), false},
		{"wrapped AppError", fmt.Errorf("wrapped: %w", NewAppError(1001, "test", "")), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAppError(tt.err); got != tt.expected {
				t.Errorf("IsAppError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetAppError(t *testing.T) {
	t.Run("AppError", func(t *testing.T) {
		original := NewAppError(2001, "test", "detail")
		got := GetAppError(original)
		if got.Code != 2001 {
			t.Errorf("expected code 2001, got %d", got.Code)
		}
	})

	t.Run("non-AppError returns ErrInternalError", func(t *testing.T) {
		got := GetAppError(errors.New("standard"))
		if got.Code != CodeInternalError {
			t.Errorf("expected code %d, got %d", CodeInternalError, got.Code)
		}
	})
}

func TestWithDetail(t *testing.T) {
	original := NewAppError(2001, "钱包不存在", "")
	withDetail := WithDetail(original, "wallet_id=W123")

	// new error has detail
	if withDetail.Detail != "wallet_id=W123" {
		t.Errorf("expected detail 'wallet_id=W123', got %q", withDetail.Detail)
	}
	if withDetail.Code != original.Code {
		t.Errorf("code should be preserved, got %d", withDetail.Code)
	}
	if withDetail.Message != original.Message {
		t.Errorf("message should be preserved, got %q", withDetail.Message)
	}

	// original is not modified
	if original.Detail != "" {
		t.Errorf("original should not be modified, detail = %q", original.Detail)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		err  *AppError
		code int
	}{
		{ErrInternalError, CodeInternalError},
		{ErrInvalidRequest, CodeInvalidRequest},
		{ErrServiceUnavailable, CodeServiceUnavailable},
		{ErrNotImplemented, CodeNotImplemented},
		{ErrWalletNotFound, CodeWalletNotFound},
		{ErrWalletExists, CodeWalletExists},
		{ErrWalletFrozen, CodeWalletFrozen},
		{ErrInvalidPassword, CodeInvalidPassword},
		{ErrInsufficientBalance, CodeInsufficientBalance},
		{ErrInvalidAmount, CodeInvalidAmount},
		{ErrTargetWalletNotFound, CodeTargetWalletNotFound},
		{ErrTransactionNotFound, CodeTransactionNotFound},
		{ErrDuplicateTransaction, CodeDuplicateTransaction},
		{ErrInvalidTransaction, CodeInvalidTransaction},
		{ErrUnauthorized, CodeUnauthorized},
		{ErrSessionExpired, CodeSessionExpired},
		{ErrInvalidToken, CodeInvalidToken},
		{ErrLockTimeout, CodeLockTimeout},
		{ErrOptimisticLock, CodeOptimisticLock},
		{ErrProcessing, CodeProcessing},
	}

	for _, tt := range tests {
		t.Run(tt.err.Message, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code %d, got %d", tt.code, tt.err.Code)
			}
		})
	}
}
