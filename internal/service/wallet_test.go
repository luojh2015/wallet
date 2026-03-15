package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/luojh/wallet/internal/domain/entity"
	dwallet "github.com/luojh/wallet/internal/domain/wallet"
	"github.com/luojh/wallet/pkg/idgen"
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

// --- mock WalletRepository ---

type mockWalletRepo struct {
	createWalletFn                func(ctx context.Context, w *entity.Wallet) error
	getWalletByIDFn               func(ctx context.Context, id string) (*entity.Wallet, error)
	updateWalletFn                func(ctx context.Context, id, name, hashPwd string) error
	updateBalanceFn               func(ctx context.Context, id string, amount int64, typ entity.TransactionType) error
	getTransactionByIdempotencyFn func(ctx context.Context, walletID, key string) (*entity.Transaction, error)
	getTransactionFn              func(ctx context.Context, id string) (*entity.Transaction, error)
	createTransactionFn           func(ctx context.Context, ts *entity.Transaction) error
	listTransactionsFn            func(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error)
}

func (m *mockWalletRepo) CreateWallet(ctx context.Context, w *entity.Wallet) error {
	if m.createWalletFn != nil {
		return m.createWalletFn(ctx, w)
	}
	return nil
}
func (m *mockWalletRepo) GetWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	if m.getWalletByIDFn != nil {
		return m.getWalletByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockWalletRepo) UpdateWallet(ctx context.Context, id, name, hashPwd string) error {
	if m.updateWalletFn != nil {
		return m.updateWalletFn(ctx, id, name, hashPwd)
	}
	return nil
}
func (m *mockWalletRepo) UpdateBalance(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
	if m.updateBalanceFn != nil {
		return m.updateBalanceFn(ctx, id, amount, typ)
	}
	return nil
}
func (m *mockWalletRepo) GetTransactionByIdempotencyKey(ctx context.Context, walletID, key string) (*entity.Transaction, error) {
	if m.getTransactionByIdempotencyFn != nil {
		return m.getTransactionByIdempotencyFn(ctx, walletID, key)
	}
	return nil, nil
}
func (m *mockWalletRepo) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	if m.getTransactionFn != nil {
		return m.getTransactionFn(ctx, id)
	}
	return nil, nil
}
func (m *mockWalletRepo) CreateTransaction(ctx context.Context, ts *entity.Transaction) error {
	if m.createTransactionFn != nil {
		return m.createTransactionFn(ctx, ts)
	}
	return nil
}
func (m *mockWalletRepo) ListTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
	if m.listTransactionsFn != nil {
		return m.listTransactionsFn(ctx, walletID, offset, limit)
	}
	return nil, 0, nil
}

// --- helper ---

func newTestWalletService() (*WalletService, *mockWalletRepo) {
	repo := &mockWalletRepo{}
	factory := idgen.NewIDFactory(&mockIDGenerator{})
	domain := dwallet.NewWallet(repo, factory)
	return NewWalletService(domain), repo
}

// --- Tests ---

func TestWalletService_CreateWallet(t *testing.T) {
	svc, _ := newTestWalletService()

	wallet, err := svc.CreateWallet(context.Background(), "test", "password")
	if err != nil {
		t.Fatalf("CreateWallet() error = %v", err)
	}
	if wallet == nil {
		t.Fatal("CreateWallet() returned nil")
	}
	if wallet.Name != "test" {
		t.Errorf("Name = %q, want %q", wallet.Name, "test")
	}
}

func TestWalletService_GetWallet(t *testing.T) {
	svc, repo := newTestWalletService()
	expected := entity.NewWallet("W1", "test", "hash")
	repo.getWalletByIDFn = func(ctx context.Context, id string) (*entity.Wallet, error) {
		return expected, nil
	}

	got, err := svc.GetWallet(context.Background(), "W1")
	if err != nil {
		t.Fatalf("GetWallet() error = %v", err)
	}
	if got.ID != "W1" {
		t.Errorf("ID = %q, want %q", got.ID, "W1")
	}
}

func TestWalletService_Deposit_Success(t *testing.T) {
	svc, _ := newTestWalletService()

	tx, err := svc.Deposit(context.Background(), "W1", "key1", 500)
	if err != nil {
		t.Fatalf("Deposit() error = %v", err)
	}
	if tx == nil {
		t.Fatal("Deposit() returned nil")
	}
}

func TestWalletService_Deposit_InvalidParams(t *testing.T) {
	svc, _ := newTestWalletService()
	ctx := context.Background()

	tests := []struct {
		name    string
		id, key string
		amount  int64
	}{
		{"empty id", "", "key", 100},
		{"empty key", "W1", "", 100},
		{"zero amount", "W1", "key", 0},
		{"negative amount", "W1", "key", -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Deposit(ctx, tt.id, tt.key, tt.amount)
			if err == nil {
				t.Error("Deposit() expected error for invalid params, got nil")
			}
		})
	}
}

func TestWalletService_Withdraw_Success(t *testing.T) {
	svc, _ := newTestWalletService()

	tx, err := svc.Withdraw(context.Background(), "W1", "key1", 300)
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if tx == nil {
		t.Fatal("Withdraw() returned nil")
	}
}

func TestWalletService_Withdraw_InvalidParams(t *testing.T) {
	svc, _ := newTestWalletService()
	ctx := context.Background()

	tests := []struct {
		name    string
		id, key string
		amount  int64
	}{
		{"empty id", "", "key", 100},
		{"empty key", "W1", "", 100},
		{"zero amount", "W1", "key", 0},
		{"negative amount", "W1", "key", -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Withdraw(ctx, tt.id, tt.key, tt.amount)
			if err == nil {
				t.Error("Withdraw() expected error for invalid params, got nil")
			}
		})
	}
}

func TestWalletService_Transfer_Success(t *testing.T) {
	svc, repo := newTestWalletService()
	fromWallet := &entity.Wallet{ID: "W1", Balance: 1000, Status: entity.WalletStatusActive}
	repo.getWalletByIDFn = func(ctx context.Context, id string) (*entity.Wallet, error) {
		return fromWallet, nil
	}

	tx, err := svc.Transfer(context.Background(), "W1", "W2", "key1", 500)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if tx == nil {
		t.Fatal("Transfer() returned nil")
	}
}

func TestWalletService_Transfer_InvalidParams(t *testing.T) {
	svc, _ := newTestWalletService()
	ctx := context.Background()

	tests := []struct {
		name          string
		from, to, key string
		amount        int64
	}{
		{"empty from", "", "W2", "key", 100},
		{"empty to", "W1", "", "key", 100},
		{"same wallet", "W1", "W1", "key", 100},
		{"empty key", "W1", "W2", "", 100},
		{"zero amount", "W1", "W2", "key", 0},
		{"negative amount", "W1", "W2", "key", -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Transfer(ctx, tt.from, tt.to, tt.key, tt.amount)
			if err == nil {
				t.Error("Transfer() expected error for invalid params, got nil")
			}
		})
	}
}

func TestWalletService_GetTransaction(t *testing.T) {
	svc, repo := newTestWalletService()
	expected := entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1")
	repo.getTransactionFn = func(ctx context.Context, id string) (*entity.Transaction, error) {
		return expected, nil
	}

	got, err := svc.GetTransaction(context.Background(), "T1")
	if err != nil {
		t.Fatalf("GetTransaction() error = %v", err)
	}
	if got.ID != "T1" {
		t.Errorf("ID = %q, want %q", got.ID, "T1")
	}
}

func TestWalletService_ListTransactions(t *testing.T) {
	svc, repo := newTestWalletService()
	txList := []*entity.Transaction{
		entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1"),
	}
	repo.listTransactionsFn = func(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
		return txList, 1, nil
	}

	list, total, err := svc.ListTransactions(context.Background(), "W1", 0, 10)
	if err != nil {
		t.Fatalf("ListTransactions() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(list) != 1 {
		t.Errorf("len = %d, want 1", len(list))
	}
}
