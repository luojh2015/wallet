package wallet

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/pkg/errors"
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

func newTestIDFactory() *idgen.IDFactory {
	return idgen.NewIDFactory(&mockIDGenerator{})
}

func newTestIDFactoryWithError(err error) *idgen.IDFactory {
	return idgen.NewIDFactory(&mockIDGenerator{err: err})
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

// --- Tests ---

func TestWallet_CreateWallet_Success(t *testing.T) {
	var created bool
	repo := &mockWalletRepo{
		createWalletFn: func(ctx context.Context, w *entity.Wallet) error {
			created = true
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	wallet, err := w.CreateWallet(context.Background(), "test", "password123")
	if err != nil {
		t.Fatalf("CreateWallet() error = %v", err)
	}
	if wallet == nil {
		t.Fatal("CreateWallet() returned nil wallet")
	}
	if wallet.Name != "test" {
		t.Errorf("Name = %q, want %q", wallet.Name, "test")
	}
	if wallet.Balance != 0 {
		t.Errorf("Balance = %d, want 0", wallet.Balance)
	}
	if wallet.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}
	if !created {
		t.Error("repo.CreateWallet should have been called")
	}
}

func TestWallet_CreateWallet_IDGenError(t *testing.T) {
	repo := &mockWalletRepo{}
	w := NewWallet(repo, newTestIDFactoryWithError(fmt.Errorf("id gen error")))

	_, err := w.CreateWallet(context.Background(), "test", "pass")
	if err == nil {
		t.Fatal("CreateWallet() expected error, got nil")
	}
}

func TestWallet_CreateWallet_RepoError(t *testing.T) {
	repo := &mockWalletRepo{
		createWalletFn: func(ctx context.Context, w *entity.Wallet) error {
			return errors.ErrWalletExists
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	_, err := w.CreateWallet(context.Background(), "test", "pass")
	if err != errors.ErrWalletExists {
		t.Errorf("CreateWallet() error = %v, want ErrWalletExists", err)
	}
}

func TestWallet_GetWallet(t *testing.T) {
	expected := entity.NewWallet("W1", "test", "hash")
	repo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			if id == "W1" {
				return expected, nil
			}
			return nil, errors.ErrWalletNotFound
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	got, err := w.GetWallet(context.Background(), "W1")
	if err != nil {
		t.Fatalf("GetWallet() error = %v", err)
	}
	if got.ID != "W1" {
		t.Errorf("ID = %q, want %q", got.ID, "W1")
	}
}

func TestWallet_UpdateWallet_WithPassword(t *testing.T) {
	var receivedHash string
	repo := &mockWalletRepo{
		updateWalletFn: func(ctx context.Context, id, name, hashPwd string) error {
			receivedHash = hashPwd
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	err := w.UpdateWallet(context.Background(), "W1", "newname", "newpass")
	if err != nil {
		t.Fatalf("UpdateWallet() error = %v", err)
	}
	if receivedHash == "" {
		t.Error("password should be hashed and passed to repo")
	}
	if receivedHash == "newpass" {
		t.Error("password should be hashed, not stored as plaintext")
	}
}

func TestWallet_UpdateWallet_WithoutPassword(t *testing.T) {
	var receivedHash string
	repo := &mockWalletRepo{
		updateWalletFn: func(ctx context.Context, id, name, hashPwd string) error {
			receivedHash = hashPwd
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	err := w.UpdateWallet(context.Background(), "W1", "newname", "")
	if err != nil {
		t.Fatalf("UpdateWallet() error = %v", err)
	}
	if receivedHash != "" {
		t.Errorf("empty password should pass empty hash, got %q", receivedHash)
	}
}

func TestWallet_Deposit_Success(t *testing.T) {
	var createdTx *entity.Transaction
	repo := &mockWalletRepo{
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			createdTx = ts
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Deposit(context.Background(), "W1", "key1", 500)
	if err != nil {
		t.Fatalf("Deposit() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusCompleted {
		t.Errorf("Status = %v, want Completed", tx.Status)
	}
	if tx.Amount != 500 {
		t.Errorf("Amount = %d, want 500", tx.Amount)
	}
	if createdTx == nil {
		t.Error("transaction should be persisted")
	}
}

func TestWallet_Deposit_Idempotent(t *testing.T) {
	existing := entity.NewTransaction("T-existing", "W1", "", 500, entity.TransactionTypeDeposit, "key1")
	existing.MarkCompleted()

	var updateBalanceCalled bool
	repo := &mockWalletRepo{
		getTransactionByIdempotencyFn: func(ctx context.Context, walletID, key string) (*entity.Transaction, error) {
			return existing, nil
		},
		updateBalanceFn: func(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
			updateBalanceCalled = true
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Deposit(context.Background(), "W1", "key1", 500)
	if err != nil {
		t.Fatalf("Deposit() error = %v", err)
	}
	if tx.ID != "T-existing" {
		t.Errorf("should return existing transaction, got ID = %q", tx.ID)
	}
	if updateBalanceCalled {
		t.Error("should not call UpdateBalance for idempotent request")
	}
}

func TestWallet_Deposit_BalanceUpdateFails(t *testing.T) {
	var createdTx *entity.Transaction
	repo := &mockWalletRepo{
		updateBalanceFn: func(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
			return errors.ErrWalletNotFound
		},
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			createdTx = ts
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Deposit(context.Background(), "W1", "key1", 500)
	if err != nil {
		t.Fatalf("Deposit() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusFailed {
		t.Errorf("Status = %v, want Failed when balance update fails", tx.Status)
	}
	if createdTx == nil {
		t.Error("failed transaction should still be persisted")
	}
}

func TestWallet_Withdraw_Success(t *testing.T) {
	repo := &mockWalletRepo{
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Withdraw(context.Background(), "W1", "key1", 300)
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusCompleted {
		t.Errorf("Status = %v, want Completed", tx.Status)
	}
}

func TestWallet_Withdraw_InsufficientBalance(t *testing.T) {
	repo := &mockWalletRepo{
		updateBalanceFn: func(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
			return errors.ErrInsufficientBalance
		},
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Withdraw(context.Background(), "W1", "key1", 9999)
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusFailed {
		t.Errorf("Status = %v, want Failed for insufficient balance", tx.Status)
	}
}

func TestWallet_Transfer_Success(t *testing.T) {
	fromWallet := &entity.Wallet{ID: "W1", Balance: 1000, Status: entity.WalletStatusActive}
	var updateCount int
	repo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			if id == "W1" {
				return fromWallet, nil
			}
			return &entity.Wallet{ID: "W2", Balance: 0, Status: entity.WalletStatusActive}, nil
		},
		updateBalanceFn: func(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
			updateCount++
			return nil
		},
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Transfer(context.Background(), "W1", "W2", "key1", 500)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusCompleted {
		t.Errorf("Status = %v, want Completed", tx.Status)
	}
	if updateCount != 2 {
		t.Errorf("UpdateBalance called %d times, want 2 (withdraw + deposit)", updateCount)
	}
}

func TestWallet_Transfer_Idempotent(t *testing.T) {
	existing := entity.NewTransaction("T-existing", "W1", "W2", 500, entity.TransactionTypeTransfer, "key1")
	existing.MarkCompleted()

	repo := &mockWalletRepo{
		getTransactionByIdempotencyFn: func(ctx context.Context, walletID, key string) (*entity.Transaction, error) {
			return existing, nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Transfer(context.Background(), "W1", "W2", "key1", 500)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if tx.ID != "T-existing" {
		t.Errorf("should return existing transaction, got ID = %q", tx.ID)
	}
}

func TestWallet_Transfer_InsufficientBalance(t *testing.T) {
	fromWallet := &entity.Wallet{ID: "W1", Balance: 100, Status: entity.WalletStatusActive}
	repo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			return fromWallet, nil
		},
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Transfer(context.Background(), "W1", "W2", "key1", 500)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusFailed {
		t.Errorf("Status = %v, want Failed for insufficient balance", tx.Status)
	}
}

func TestWallet_Transfer_TargetDepositFails_Rollback(t *testing.T) {
	fromWallet := &entity.Wallet{ID: "W1", Balance: 1000, Status: entity.WalletStatusActive}
	var updateCalls []struct {
		id  string
		typ entity.TransactionType
	}
	callCount := 0
	repo := &mockWalletRepo{
		getWalletByIDFn: func(ctx context.Context, id string) (*entity.Wallet, error) {
			return fromWallet, nil
		},
		updateBalanceFn: func(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
			updateCalls = append(updateCalls, struct {
				id  string
				typ entity.TransactionType
			}{id, typ})
			callCount++
			if callCount == 2 {
				// second call (deposit to target) fails
				return errors.ErrWalletNotFound
			}
			return nil
		},
		createTransactionFn: func(ctx context.Context, ts *entity.Transaction) error {
			return nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	tx, err := w.Transfer(context.Background(), "W1", "W2", "key1", 500)
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if tx.Status != entity.TransactionStatusFailed {
		t.Errorf("Status = %v, want Failed", tx.Status)
	}
	// should have 3 calls: withdraw from source, deposit to target (fails), rollback to source
	if len(updateCalls) != 3 {
		t.Fatalf("UpdateBalance called %d times, want 3 (withdraw + failed deposit + rollback)", len(updateCalls))
	}
	// third call should be a rollback deposit to source
	if updateCalls[2].id != "W1" || updateCalls[2].typ != entity.TransactionTypeDeposit {
		t.Errorf("rollback call: id=%q typ=%v, want id=W1 typ=Deposit", updateCalls[2].id, updateCalls[2].typ)
	}
}

func TestWallet_GetTransaction(t *testing.T) {
	expected := entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1")
	repo := &mockWalletRepo{
		getTransactionFn: func(ctx context.Context, id string) (*entity.Transaction, error) {
			return expected, nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	got, err := w.GetTransaction(context.Background(), "T1")
	if err != nil {
		t.Fatalf("GetTransaction() error = %v", err)
	}
	if got.ID != "T1" {
		t.Errorf("ID = %q, want %q", got.ID, "T1")
	}
}

func TestWallet_ListTransactions(t *testing.T) {
	txList := []*entity.Transaction{
		entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1"),
	}
	repo := &mockWalletRepo{
		listTransactionsFn: func(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
			return txList, 1, nil
		},
	}
	w := NewWallet(repo, newTestIDFactory())

	list, total, err := w.ListTransactions(context.Background(), "W1", 0, 10)
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
