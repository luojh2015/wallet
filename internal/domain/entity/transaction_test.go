package entity

import (
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {
	before := time.Now()
	tx := NewTransaction("T1", "W1", "W2", 1000, TransactionTypeTransfer, "key-001")
	after := time.Now()

	if tx.ID != "T1" {
		t.Errorf("ID = %q, want %q", tx.ID, "T1")
	}
	if tx.FromWalletID != "W1" {
		t.Errorf("FromWalletID = %q, want %q", tx.FromWalletID, "W1")
	}
	if tx.ToWalletID != "W2" {
		t.Errorf("ToWalletID = %q, want %q", tx.ToWalletID, "W2")
	}
	if tx.Amount != 1000 {
		t.Errorf("Amount = %d, want 1000", tx.Amount)
	}
	if tx.Type != TransactionTypeTransfer {
		t.Errorf("Type = %v, want Transfer", tx.Type)
	}
	if tx.Status != TransactionStatusPending {
		t.Errorf("Status = %v, want Pending", tx.Status)
	}
	if tx.IdempotencyKey != "key-001" {
		t.Errorf("IdempotencyKey = %q, want %q", tx.IdempotencyKey, "key-001")
	}
	if tx.CompletedAt != nil {
		t.Errorf("CompletedAt should be nil, got %v", tx.CompletedAt)
	}
	if tx.CreatedAt.Before(before) || tx.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not in range [%v, %v]", tx.CreatedAt, before, after)
	}
}

func TestTransactionType_String(t *testing.T) {
	tests := []struct {
		typ  TransactionType
		want string
	}{
		{TransactionTypeTransfer, "transfer"},
		{TransactionTypeDeposit, "deposit"},
		{TransactionTypeWithdraw, "withdraw"},
		{TransactionType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("TransactionType(%d).String() = %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestTransactionStatus_String(t *testing.T) {
	tests := []struct {
		status TransactionStatus
		want   string
	}{
		{TransactionStatusPending, "pending"},
		{TransactionStatusCompleted, "completed"},
		{TransactionStatusFailed, "failed"},
		{TransactionStatusCancelled, "cancelled"},
		{TransactionStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("TransactionStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestTransaction_MarkCompleted(t *testing.T) {
	tx := NewTransaction("T1", "W1", "W2", 100, TransactionTypeTransfer, "key")
	before := time.Now()
	tx.MarkCompleted()
	after := time.Now()

	if tx.Status != TransactionStatusCompleted {
		t.Errorf("Status = %v, want Completed", tx.Status)
	}
	if tx.CompletedAt == nil {
		t.Fatal("CompletedAt should not be nil after MarkCompleted")
	}
	if tx.CompletedAt.Before(before) || tx.CompletedAt.After(after) {
		t.Errorf("CompletedAt %v not in range [%v, %v]", *tx.CompletedAt, before, after)
	}
}

func TestTransaction_MarkFailed(t *testing.T) {
	tx := NewTransaction("T1", "W1", "W2", 100, TransactionTypeTransfer, "key")
	tx.MarkFailed()

	if tx.Status != TransactionStatusFailed {
		t.Errorf("Status = %v, want Failed", tx.Status)
	}
	if tx.CompletedAt == nil {
		t.Fatal("CompletedAt should not be nil after MarkFailed")
	}
}

func TestTransaction_MarkCancelled(t *testing.T) {
	tx := NewTransaction("T1", "W1", "W2", 100, TransactionTypeTransfer, "key")
	tx.MarkCancelled()

	if tx.Status != TransactionStatusCancelled {
		t.Errorf("Status = %v, want Cancelled", tx.Status)
	}
	if tx.CompletedAt == nil {
		t.Fatal("CompletedAt should not be nil after MarkCancelled")
	}
}
