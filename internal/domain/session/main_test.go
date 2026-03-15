package session

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/pkg/errors"
	"github.com/luojh/wallet/pkg/idgen"
	"github.com/luojh/wallet/pkg/pwd"
	"go.uber.org/fx"
)

// --- mock IDGenerator ---

type mockIDGenerator struct {
	counter atomic.Int64
	err     error
}

func (m *mockIDGenerator) Generate() (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.counter.Add(1), nil
}
func (m *mockIDGenerator) GenerateString() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return fmt.Sprintf("%d", m.counter.Add(1)), nil
}
func (m *mockIDGenerator) Clone() idgen.IDGenerator {
	return &mockIDGenerator{err: m.err}
}
func (m *mockIDGenerator) Parse(id int64) idgen.ParsedID {
	return idgen.ParsedID{}
}

// --- mock SessionRepository ---

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

func newTestSessionDomain(repo *mockSessionRepo, genErr error) *Session {
	gen := &mockIDGenerator{err: genErr}
	factory := idgen.NewIDFactory(gen)
	cfg := config.DefaultConfig()
	cfg.Session.TTL = time.Hour // large TTL to avoid cleanup during tests
	lc := &mockLifecycle{}
	return NewSessionDomain(lc, repo, factory, cfg)
}

// --- Tests ---

func TestSession_CreateSession_Success(t *testing.T) {
	var created bool
	repo := &mockSessionRepo{
		createFn: func(ctx context.Context, session *entity.Session) error {
			created = true
			return nil
		},
	}
	s := newTestSessionDomain(repo, nil)

	session, err := s.CreateSession(context.Background(), "W1")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session == nil {
		t.Fatal("CreateSession() returned nil")
	}
	if session.WalletID != "W1" {
		t.Errorf("WalletID = %q, want %q", session.WalletID, "W1")
	}
	if session.Token == "" {
		t.Error("Token should not be empty")
	}
	if !session.IsValid {
		t.Error("IsValid should be true")
	}
	if !created {
		t.Error("repo.Create should have been called")
	}
}

func TestSession_CreateSession_CustomTTL(t *testing.T) {
	repo := &mockSessionRepo{}
	s := newTestSessionDomain(repo, nil)

	before := time.Now()
	session, err := s.CreateSession(context.Background(), "W1", 2*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	expectedExpiry := before.Add(2 * time.Hour)
	if session.ExpiresAt.Before(expectedExpiry.Add(-2*time.Second)) || session.ExpiresAt.After(expectedExpiry.Add(2*time.Second)) {
		t.Errorf("ExpiresAt = %v, expected around %v", session.ExpiresAt, expectedExpiry)
	}
}

func TestSession_CreateSession_IDGenError(t *testing.T) {
	repo := &mockSessionRepo{}
	s := newTestSessionDomain(repo, fmt.Errorf("id gen error"))

	_, err := s.CreateSession(context.Background(), "W1")
	if err == nil {
		t.Fatal("CreateSession() expected error, got nil")
	}
}

func TestSession_CreateSession_RepoError(t *testing.T) {
	repo := &mockSessionRepo{
		createFn: func(ctx context.Context, session *entity.Session) error {
			return fmt.Errorf("repo error")
		},
	}
	s := newTestSessionDomain(repo, nil)

	_, err := s.CreateSession(context.Background(), "W1")
	if err == nil {
		t.Fatal("CreateSession() expected error, got nil")
	}
}

func TestSession_ValidateSession_Success(t *testing.T) {
	validSession := &entity.Session{
		ID:        "S1",
		WalletID:  "W1",
		Token:     "valid-token",
		IsValid:   true,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return validSession, nil
		},
	}
	s := newTestSessionDomain(repo, nil)

	session, err := s.ValidateSession(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}
	if session.WalletID != "W1" {
		t.Errorf("WalletID = %q, want %q", session.WalletID, "W1")
	}
}

func TestSession_ValidateSession_Expired(t *testing.T) {
	expiredSession := &entity.Session{
		ID:        "S1",
		WalletID:  "W1",
		Token:     "expired-token",
		IsValid:   true,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	repo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return expiredSession, nil
		},
	}
	s := newTestSessionDomain(repo, nil)

	_, err := s.ValidateSession(context.Background(), "expired-token")
	if err != errors.ErrSessionExpired {
		t.Errorf("ValidateSession() error = %v, want ErrSessionExpired", err)
	}
}

func TestSession_ValidateSession_Invalidated(t *testing.T) {
	invalidSession := &entity.Session{
		ID:        "S1",
		WalletID:  "W1",
		Token:     "invalid-token",
		IsValid:   false,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return invalidSession, nil
		},
	}
	s := newTestSessionDomain(repo, nil)

	_, err := s.ValidateSession(context.Background(), "invalid-token")
	if err != errors.ErrSessionExpired {
		t.Errorf("ValidateSession() error = %v, want ErrSessionExpired", err)
	}
}

func TestSession_ValidateSession_NotFound(t *testing.T) {
	repo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return nil, errors.ErrInvalidToken
		},
	}
	s := newTestSessionDomain(repo, nil)

	_, err := s.ValidateSession(context.Background(), "unknown")
	if err != errors.ErrInvalidToken {
		t.Errorf("ValidateSession() error = %v, want ErrInvalidToken", err)
	}
}

func TestSession_InvalidateSession(t *testing.T) {
	var invalidatedToken string
	repo := &mockSessionRepo{
		invalidateFn: func(ctx context.Context, token string) error {
			invalidatedToken = token
			return nil
		},
	}
	s := newTestSessionDomain(repo, nil)

	err := s.InvalidateSession(context.Background(), "my-token")
	if err != nil {
		t.Fatalf("InvalidateSession() error = %v", err)
	}
	if invalidatedToken != "my-token" {
		t.Errorf("invalidated token = %q, want %q", invalidatedToken, "my-token")
	}
}

func TestSession_GetWalletID_Success(t *testing.T) {
	validSession := &entity.Session{
		ID:        "S1",
		WalletID:  "W1",
		Token:     "token",
		IsValid:   true,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	repo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return validSession, nil
		},
	}
	s := newTestSessionDomain(repo, nil)

	walletID, err := s.GetWalletID(context.Background(), "token")
	if err != nil {
		t.Fatalf("GetWalletID() error = %v", err)
	}
	if walletID != "W1" {
		t.Errorf("walletID = %q, want %q", walletID, "W1")
	}
}

func TestSession_GetWalletID_InvalidSession(t *testing.T) {
	repo := &mockSessionRepo{
		getByTokenFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return nil, errors.ErrInvalidToken
		},
	}
	s := newTestSessionDomain(repo, nil)

	_, err := s.GetWalletID(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("GetWalletID() expected error, got nil")
	}
}

func TestSession_VerifyPassword_Success(t *testing.T) {
	repo := &mockSessionRepo{}
	s := newTestSessionDomain(repo, nil)

	// VerifyPassword internally uses pwd.CheckPasswordHash
	// Use pwd package to generate a real hash for testing
	passwordHash, err := pwd.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("pwd.HashPassword() error = %v", err)
	}

	err = s.VerifyPassword(context.Background(), passwordHash, "correct-password")
	if err != nil {
		t.Errorf("VerifyPassword() with correct password, error = %v, want nil", err)
	}
}

func TestSession_VerifyPassword_Wrong(t *testing.T) {
	repo := &mockSessionRepo{}
	s := newTestSessionDomain(repo, nil)

	err := s.VerifyPassword(context.Background(), "some-hash", "wrong-password")
	if err != errors.ErrInvalidPassword {
		t.Errorf("VerifyPassword() error = %v, want ErrInvalidPassword", err)
	}
}
