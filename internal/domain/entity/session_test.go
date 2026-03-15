package entity

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	before := time.Now()
	s, err := NewSession("S1", "W1", time.Hour)
	after := time.Now()

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	if s.ID != "S1" {
		t.Errorf("ID = %q, want %q", s.ID, "S1")
	}
	if s.WalletID != "W1" {
		t.Errorf("WalletID = %q, want %q", s.WalletID, "W1")
	}
	if len(s.Token) != 64 {
		t.Errorf("Token length = %d, want 64 (32 bytes hex encoded)", len(s.Token))
	}
	if !s.IsValid {
		t.Error("IsValid should be true for new session")
	}
	if s.CreatedAt.Before(before) || s.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not in range [%v, %v]", s.CreatedAt, before, after)
	}
	// ExpiresAt should be roughly CreatedAt + 1 hour
	expectedExpiry := s.CreatedAt.Add(time.Hour)
	if s.ExpiresAt.Before(expectedExpiry.Add(-time.Second)) || s.ExpiresAt.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("ExpiresAt %v not close to expected %v", s.ExpiresAt, expectedExpiry)
	}
}

func TestNewSession_UniqueTokens(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		s, err := NewSession("S1", "W1", time.Hour)
		if err != nil {
			t.Fatalf("NewSession() error = %v", err)
		}
		if tokens[s.Token] {
			t.Fatalf("duplicate token generated: %s", s.Token)
		}
		tokens[s.Token] = true
	}
}

func TestSession_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		s := &Session{ExpiresAt: time.Now().Add(time.Hour)}
		if s.IsExpired() {
			t.Error("session should not be expired")
		}
	})

	t.Run("expired", func(t *testing.T) {
		s := &Session{ExpiresAt: time.Now().Add(-time.Hour)}
		if !s.IsExpired() {
			t.Error("session should be expired")
		}
	})
}

func TestSession_IsValidSession(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
		expired bool
		want    bool
	}{
		{"valid and fresh", true, false, true},
		{"valid but expired", true, true, false},
		{"invalid but fresh", false, false, false},
		{"invalid and expired", false, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiry := time.Now().Add(time.Hour)
			if tt.expired {
				expiry = time.Now().Add(-time.Hour)
			}
			s := &Session{
				IsValid:   tt.isValid,
				ExpiresAt: expiry,
			}
			if got := s.IsValidSession(); got != tt.want {
				t.Errorf("IsValidSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_Invalidate(t *testing.T) {
	s := &Session{IsValid: true}
	s.Invalidate()
	if s.IsValid {
		t.Error("IsValid should be false after Invalidate()")
	}
}
