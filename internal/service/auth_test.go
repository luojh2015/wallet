package service

import (
	"context"
	"testing"
	"time"

	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/internal/domain/entity"
	dsession "github.com/luojh/wallet/internal/domain/session"
	"github.com/luojh/wallet/pkg/errors"
	"github.com/luojh/wallet/pkg/idgen"
	"github.com/luojh/wallet/pkg/pwd"
	"go.uber.org/fx"
)

// --- mock SessionRepository (for auth tests) ---

type mockSessionRepo struct {
	createFn             func(ctx context.Context, session *entity.Session) error
	getByTokenFn         func(ctx context.Context, token string) (*entity.Session, error)
	getByWalletIDFn      func(ctx context.Context, id string) (*entity.Session, error)
	invalidateFn         func(ctx context.Context, token string) error
	invalidateByWalletFn func(ctx context.Context, id string) error
	cleanupFn            func(ctx context.Context) error
}

func (m *mockSessionRepo) Create(ctx context.Context, session *entity.Session) error {
	if m.createFn != nil {
		return m.createFn(ctx, session)
	}
	return nil
}
func (m *mockSessionRepo) GetByToken(ctx context.Context, token string) (*entity.Session, error) {
	if m.getByTokenFn != nil {
		return m.getByTokenFn(ctx, token)
	}
	return nil, nil
}
func (m *mockSessionRepo) GetByWalletID(ctx context.Context, id string) (*entity.Session, error) {
	if m.getByWalletIDFn != nil {
		return m.getByWalletIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockSessionRepo) Invalidate(ctx context.Context, token string) error {
	if m.invalidateFn != nil {
		return m.invalidateFn(ctx, token)
	}
	return nil
}
func (m *mockSessionRepo) InvalidateByWalletID(ctx context.Context, id string) error {
	if m.invalidateByWalletFn != nil {
		return m.invalidateByWalletFn(ctx, id)
	}
	return nil
}
func (m *mockSessionRepo) Cleanup(ctx context.Context) error {
	if m.cleanupFn != nil {
		return m.cleanupFn(ctx)
	}
	return nil
}

// --- mock fx.Lifecycle ---

type mockLifecycle struct {
	hooks []fx.Hook
}

func (m *mockLifecycle) Append(hook fx.Hook) {
	m.hooks = append(m.hooks, hook)
}

// --- helper ---

func newTestAuthService(walletRepo *mockWalletRepo, sessionRepo *mockSessionRepo) IAuthService {
	gen := &mockIDGenerator{}
	factory := idgen.NewIDFactory(gen)
	cfg := config.DefaultConfig()
	cfg.Session.TTL = time.Hour
	lc := &mockLifecycle{}
	sessionDomain := dsession.NewSessionDomain(lc, sessionRepo, factory, cfg)
	return NewAuthService(sessionDomain, walletRepo, factory)
}

// --- Tests ---

func TestAuthService_Login_Success(t *testing.T) {
	passwordHash, err := pwd.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	walletRepo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			return &entity.Wallet{
				ID:           "W1",
				Name:         "test",
				PasswordHash: passwordHash,
				Status:       entity.WalletStatusActive,
			}, nil
		},
	}
	sessionRepo := &mockSessionRepo{}
	svc := newTestAuthService(walletRepo, sessionRepo)

	session, err := svc.Login(context.Background(), "W1", "correct-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if session == nil {
		t.Fatal("Login() returned nil session")
	}
	if session.Token == "" {
		t.Error("session.Token should not be empty")
	}
	if session.WalletID != "W1" {
		t.Errorf("session.WalletID = %q, want %q", session.WalletID, "W1")
	}
}

func TestAuthService_Login_WalletNotFound(t *testing.T) {
	walletRepo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			return nil, errors.ErrWalletNotFound
		},
	}
	sessionRepo := &mockSessionRepo{}
	svc := newTestAuthService(walletRepo, sessionRepo)

	_, err := svc.Login(context.Background(), "W999", "password")
	if err != errors.ErrWalletNotFound {
		t.Errorf("Login() error = %v, want ErrWalletNotFound", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	passwordHash, _ := pwd.HashPassword("correct-password")

	walletRepo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			return &entity.Wallet{
				ID:           "W1",
				PasswordHash: passwordHash,
				Status:       entity.WalletStatusActive,
			}, nil
		},
	}
	sessionRepo := &mockSessionRepo{}
	svc := newTestAuthService(walletRepo, sessionRepo)

	_, err := svc.Login(context.Background(), "W1", "wrong-password")
	if err != errors.ErrInvalidPassword {
		t.Errorf("Login() error = %v, want ErrInvalidPassword", err)
	}
}

func TestAuthService_Login_FrozenWallet(t *testing.T) {
	passwordHash, _ := pwd.HashPassword("password")

	walletRepo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			return &entity.Wallet{
				ID:           "W1",
				PasswordHash: passwordHash,
				Status:       entity.WalletStatusFrozen,
			}, nil
		},
	}
	sessionRepo := &mockSessionRepo{}
	svc := newTestAuthService(walletRepo, sessionRepo)

	_, err := svc.Login(context.Background(), "W1", "password")
	if err != errors.ErrWalletFrozen {
		t.Errorf("Login() error = %v, want ErrWalletFrozen", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	var invalidatedToken string
	walletRepo := &mockWalletRepo{}
	sessionRepo := &mockSessionRepo{
		invalidateFn: func(ctx context.Context, token string) error {
			invalidatedToken = token
			return nil
		},
	}
	svc := newTestAuthService(walletRepo, sessionRepo)

	err := svc.Logout(context.Background(), "my-token")
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if invalidatedToken != "my-token" {
		t.Errorf("invalidated token = %q, want %q", invalidatedToken, "my-token")
	}
}

func TestAuthService_ValidateSession(t *testing.T) {
	validSession := &entity.Session{
		ID:        "S1",
		WalletID:  "W1",
		Token:     "valid-token",
		IsValid:   true,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	walletRepo := &mockWalletRepo{}
	sessionRepo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return validSession, nil
		},
	}
	svc := newTestAuthService(walletRepo, sessionRepo)

	session, err := svc.ValidateSession(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}
	if session.WalletID != "W1" {
		t.Errorf("WalletID = %q, want %q", session.WalletID, "W1")
	}
}

func TestAuthService_GetWalletID(t *testing.T) {
	validSession := &entity.Session{
		ID:        "S1",
		WalletID:  "W1",
		Token:     "token",
		IsValid:   true,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	walletRepo := &mockWalletRepo{}
	sessionRepo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return validSession, nil
		},
	}
	svc := newTestAuthService(walletRepo, sessionRepo)

	walletID, err := svc.GetWalletID(context.Background(), "token")
	if err != nil {
		t.Fatalf("GetWalletID() error = %v", err)
	}
	if walletID != "W1" {
		t.Errorf("walletID = %q, want %q", walletID, "W1")
	}
}
