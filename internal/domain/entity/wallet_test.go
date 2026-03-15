package entity

import (
	"testing"
	"time"
)

func TestNewWallet(t *testing.T) {
	before := time.Now()
	w := NewWallet("W123", "test-wallet", "hash123")
	after := time.Now()

	if w.ID != "W123" {
		t.Errorf("ID = %q, want %q", w.ID, "W123")
	}
	if w.Name != "test-wallet" {
		t.Errorf("Name = %q, want %q", w.Name, "test-wallet")
	}
	if w.PasswordHash != "hash123" {
		t.Errorf("PasswordHash = %q, want %q", w.PasswordHash, "hash123")
	}
	if w.Balance != 0 {
		t.Errorf("Balance = %d, want 0", w.Balance)
	}
	if w.Status != WalletStatusActive {
		t.Errorf("Status = %v, want Active", w.Status)
	}
	if w.Version != 1 {
		t.Errorf("Version = %d, want 1", w.Version)
	}
	if w.CreatedAt.Before(before) || w.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not in range [%v, %v]", w.CreatedAt, before, after)
	}
	if w.UpdatedAt.Before(before) || w.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt %v not in range [%v, %v]", w.UpdatedAt, before, after)
	}
}

func TestWalletStatus_String(t *testing.T) {
	tests := []struct {
		status WalletStatus
		want   string
	}{
		{WalletStatusActive, "active"},
		{WalletStatusFrozen, "frozen"},
		{WalletStatusClosed, "closed"},
		{WalletStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("WalletStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestWallet_CanTransfer(t *testing.T) {
	tests := []struct {
		status WalletStatus
		want   bool
	}{
		{WalletStatusActive, true},
		{WalletStatusFrozen, false},
		{WalletStatusClosed, false},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			w := &Wallet{Status: tt.status}
			if got := w.CanTransfer(); got != tt.want {
				t.Errorf("CanTransfer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWallet_HasSufficientBalance(t *testing.T) {
	tests := []struct {
		name    string
		balance int64
		amount  int64
		want    bool
	}{
		{"exact", 100, 100, true},
		{"more than enough", 200, 100, true},
		{"insufficient", 50, 100, false},
		{"zero balance zero amount", 0, 0, true},
		{"zero balance positive amount", 0, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Wallet{Balance: tt.balance}
			if got := w.HasSufficientBalance(tt.amount); got != tt.want {
				t.Errorf("HasSufficientBalance(%d) = %v, want %v (balance=%d)", tt.amount, got, tt.want, tt.balance)
			}
		})
	}
}
