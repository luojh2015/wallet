package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/pkg/errors"
)

func newTestSession(id, walletID, token string, ttl time.Duration, valid bool) *entity.Session {
	now := time.Now()
	return &entity.Session{
		ID:        id,
		WalletID:  walletID,
		Token:     token,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
		IsValid:   valid,
	}
}

func TestSessionRepo_CreateAndGetByToken(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	s := newTestSession("S1", "W1", "token-abc", time.Hour, true)
	if err := repo.Create(ctx, s); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByToken(ctx, "token-abc")
	if err != nil {
		t.Fatalf("GetByToken() error = %v", err)
	}
	if got.ID != "S1" {
		t.Errorf("ID = %q, want %q", got.ID, "S1")
	}
	if got.WalletID != "W1" {
		t.Errorf("WalletID = %q, want %q", got.WalletID, "W1")
	}
}

func TestSessionRepo_GetByToken_NotFound(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	_, err := repo.GetByToken(ctx, "nonexistent")
	if err != errors.ErrInvalidToken {
		t.Errorf("GetByToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestSessionRepo_GetByWalletID(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	s := newTestSession("S1", "W1", "token-abc", time.Hour, true)
	repo.Create(ctx, s)

	got, err := repo.GetByWalletID(ctx, "S1")
	if err != nil {
		t.Fatalf("GetByWalletID() error = %v", err)
	}
	if got.WalletID != "W1" {
		t.Errorf("WalletID = %q, want %q", got.WalletID, "W1")
	}
}

func TestSessionRepo_GetByWalletID_NotFound(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	_, err := repo.GetByWalletID(ctx, "nonexistent")
	if err != errors.ErrInvalidToken {
		t.Errorf("GetByWalletID() error = %v, want ErrInvalidToken", err)
	}
}

func TestSessionRepo_GetReturnsCopy(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	s := newTestSession("S1", "W1", "token-abc", time.Hour, true)
	repo.Create(ctx, s)

	got, _ := repo.GetByToken(ctx, "token-abc")
	got.WalletID = "modified"

	got2, _ := repo.GetByToken(ctx, "token-abc")
	if got2.WalletID != "W1" {
		t.Errorf("store should be unaffected by modifying returned copy, WalletID = %q", got2.WalletID)
	}
}

func TestSessionRepo_Invalidate(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	s := newTestSession("S1", "W1", "token-abc", time.Hour, true)
	repo.Create(ctx, s)

	err := repo.Invalidate(ctx, "token-abc")
	if err != nil {
		t.Fatalf("Invalidate() error = %v", err)
	}

	got, _ := repo.GetByToken(ctx, "token-abc")
	if got.IsValid {
		t.Error("IsValid should be false after Invalidate")
	}
}

func TestSessionRepo_Invalidate_NotExist(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	err := repo.Invalidate(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Invalidate() for nonexistent token should return nil, got %v", err)
	}
}

func TestSessionRepo_InvalidateByWalletID(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	s1 := newTestSession("S1", "W1", "token1", time.Hour, true)
	s2 := newTestSession("S2", "W1", "token2", time.Hour, true)
	s3 := newTestSession("S3", "W2", "token3", time.Hour, true)
	repo.Create(ctx, s1)
	repo.Create(ctx, s2)
	repo.Create(ctx, s3)

	err := repo.InvalidateByWalletID(ctx, "W1")
	if err != nil {
		t.Fatalf("InvalidateByWalletID() error = %v", err)
	}

	got1, _ := repo.GetByToken(ctx, "token1")
	if got1.IsValid {
		t.Error("session1 should be invalidated")
	}
	got2, _ := repo.GetByToken(ctx, "token2")
	if got2.IsValid {
		t.Error("session2 should be invalidated")
	}
	got3, _ := repo.GetByToken(ctx, "token3")
	if !got3.IsValid {
		t.Error("session3 (different wallet) should remain valid")
	}
}

func TestSessionRepo_Cleanup(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	// expired session
	expired := newTestSession("S1", "W1", "token-expired", -time.Hour, true)
	repo.Create(ctx, expired)

	// invalidated session
	invalid := newTestSession("S2", "W1", "token-invalid", time.Hour, false)
	repo.Create(ctx, invalid)

	// valid session
	valid := newTestSession("S3", "W2", "token-valid", time.Hour, true)
	repo.Create(ctx, valid)

	err := repo.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// expired should be removed
	_, err = repo.GetByToken(ctx, "token-expired")
	if err != errors.ErrInvalidToken {
		t.Error("expired session should be cleaned up")
	}

	// invalidated should be removed
	_, err = repo.GetByToken(ctx, "token-invalid")
	if err != errors.ErrInvalidToken {
		t.Error("invalidated session should be cleaned up")
	}

	// valid should remain
	got, err := repo.GetByToken(ctx, "token-valid")
	if err != nil {
		t.Fatalf("valid session should remain after cleanup, error = %v", err)
	}
	if got.ID != "S3" {
		t.Errorf("valid session ID = %q, want %q", got.ID, "S3")
	}
}

func TestSessionRepo_Concurrent(t *testing.T) {
	repo := NewSession()
	ctx := context.Background()

	var wg sync.WaitGroup
	// concurrent creates
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token := "token-" + string(rune('A'+idx%26)) + string(rune('0'+idx/26))
			s := newTestSession("S"+token, "W1", token, time.Hour, true)
			repo.Create(ctx, s)
		}(i)
	}
	// concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// may or may not find, just should not panic
			repo.GetByToken(ctx, "token-A0")
		}()
	}
	wg.Wait()
}
