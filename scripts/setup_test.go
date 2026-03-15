package scripts

import (
	"context"
	"testing"
	"time"

	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/domain/session"
	"github.com/luojh/wallet/internal/domain/wallet"
	"github.com/luojh/wallet/internal/repository"
	"github.com/luojh/wallet/internal/repository/memory"
	"github.com/luojh/wallet/internal/service"
	"github.com/luojh/wallet/pkg/idgen"
	"github.com/luojh/wallet/pkg/idgen/snowflake"
	"github.com/luojh/wallet/pkg/pwd"
	"go.uber.org/fx"
)

// --- mock fx.Lifecycle ---

type mockLifecycle struct {
	hooks []fx.Hook
}

func (m *mockLifecycle) Append(hook fx.Hook) {
	m.hooks = append(m.hooks, hook)
}

func (m *mockLifecycle) stop() {
	ctx := context.Background()
	for _, h := range m.hooks {
		if h.OnStop != nil {
			_ = h.OnStop(ctx)
		}
	}
}

// --- testEnv ---

type testEnv struct {
	walletService *service.WalletService
	authService   service.IAuthService
	walletRepo    repository.IWalletRepository
	sessionRepo   repository.ISessionRepository
	idFactory     *idgen.IDFactory
	lifecycle     *mockLifecycle
}

func newTestEnv(t *testing.T) *testEnv {
	return newTestEnvWithSessionTTL(t, time.Hour)
}

func newTestEnvWithSessionTTL(t *testing.T, ttl time.Duration) *testEnv {
	t.Helper()

	gen, err := snowflake.NewSnowflakeGenerator(0)
	if err != nil {
		t.Fatalf("NewSnowflakeGenerator: %v", err)
	}
	factory := idgen.NewIDFactory(gen)

	walletRepo := memory.NewWallet()
	sessionRepo := memory.NewSession()

	cfg := config.DefaultConfig()
	cfg.Session.TTL = ttl

	lc := &mockLifecycle{}

	domainWallet := wallet.NewWallet(walletRepo, factory)
	domainSession := session.NewSessionDomain(lc, sessionRepo, factory, cfg)

	walletSvc := service.NewWalletService(domainWallet)
	authSvc := service.NewAuthService(domainSession, walletRepo, factory)

	env := &testEnv{
		walletService: walletSvc,
		authService:   authSvc,
		walletRepo:    walletRepo,
		sessionRepo:   sessionRepo,
		idFactory:     factory,
		lifecycle:     lc,
	}

	t.Cleanup(func() { env.lifecycle.stop() })

	return env
}

// --- helpers ---

func mustCreateWallet(t *testing.T, env *testEnv, name, password string) *entity.Wallet {
	t.Helper()
	ctx := context.Background()
	w, err := env.walletService.CreateWallet(ctx, name, password)
	if err != nil {
		t.Fatalf("CreateWallet(%q): %v", name, err)
	}
	return w
}

func mustDeposit(t *testing.T, env *testEnv, walletID, key string, amount int64) *entity.Transaction {
	t.Helper()
	ctx := context.Background()
	tx, err := env.walletService.Deposit(ctx, walletID, key, amount)
	if err != nil {
		t.Fatalf("Deposit(%q, %q, %d): %v", walletID, key, amount, err)
	}
	return tx
}

func mustCreateFrozenWallet(t *testing.T, env *testEnv, name, password string) *entity.Wallet {
	t.Helper()
	ctx := context.Background()

	id, err := env.idFactory.GenWalletID()
	if err != nil {
		t.Fatalf("GenWalletID: %v", err)
	}
	hash, err := pwd.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	w := entity.NewWallet(id, name, hash)
	w.Status = entity.WalletStatusFrozen

	if err := env.walletRepo.CreateWallet(ctx, w); err != nil {
		t.Fatalf("CreateWallet (frozen): %v", err)
	}
	return w
}

func assertBalance(t *testing.T, env *testEnv, walletID string, expected int64) {
	t.Helper()
	ctx := context.Background()
	w, err := env.walletService.GetWallet(ctx, walletID)
	if err != nil {
		t.Fatalf("GetWallet(%q): %v", walletID, err)
	}
	if w.Balance != expected {
		t.Errorf("wallet %q balance = %d, want %d", walletID, w.Balance, expected)
	}
}

func getBalance(t *testing.T, env *testEnv, walletID string) int64 {
	t.Helper()
	ctx := context.Background()
	w, err := env.walletService.GetWallet(ctx, walletID)
	if err != nil {
		t.Fatalf("GetWallet(%q): %v", walletID, err)
	}
	return w.Balance
}
