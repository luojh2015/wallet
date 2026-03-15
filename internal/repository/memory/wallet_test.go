package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/pkg/errors"
)

func TestWalletStore_CreateAndGet(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	if err := repo.CreateWallet(ctx, w); err != nil {
		t.Fatalf("CreateWallet() error = %v", err)
	}

	got, err := repo.GetWalletByID(ctx, "W1")
	if err != nil {
		t.Fatalf("GetWalletByID() error = %v", err)
	}
	if got.ID != "W1" {
		t.Errorf("ID = %q, want %q", got.ID, "W1")
	}
	if got.Name != "test" {
		t.Errorf("Name = %q, want %q", got.Name, "test")
	}
}

func TestWalletStore_CreateDuplicate(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	repo.CreateWallet(ctx, w)

	w2 := entity.NewWallet("W1", "test2", "hash2")
	err := repo.CreateWallet(ctx, w2)
	if err != errors.ErrWalletExists {
		t.Errorf("CreateWallet() error = %v, want ErrWalletExists", err)
	}
}

func TestWalletStore_GetNotFound(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	_, err := repo.GetWalletByID(ctx, "W999")
	if err != errors.ErrWalletNotFound {
		t.Errorf("GetWalletByID() error = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletStore_GetReturnsCopy(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "original", "hash")
	repo.CreateWallet(ctx, w)

	got, _ := repo.GetWalletByID(ctx, "W1")
	got.Name = "modified"

	got2, _ := repo.GetWalletByID(ctx, "W1")
	if got2.Name != "original" {
		t.Errorf("store should be unaffected by modifying returned copy, Name = %q", got2.Name)
	}
}

func TestWalletStore_UpdateWallet(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "original", "hash")
	repo.CreateWallet(ctx, w)

	err := repo.UpdateWallet(ctx, "W1", "updated", "newhash")
	if err != nil {
		t.Fatalf("UpdateWallet() error = %v", err)
	}

	got, _ := repo.GetWalletByID(ctx, "W1")
	if got.Name != "updated" {
		t.Errorf("Name = %q, want %q", got.Name, "updated")
	}
	if got.PasswordHash != "newhash" {
		t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, "newhash")
	}
	if got.Version != 2 {
		t.Errorf("Version = %d, want 2", got.Version)
	}
}

func TestWalletStore_UpdateNotFound(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	err := repo.UpdateWallet(ctx, "W999", "name", "hash")
	if err != errors.ErrWalletNotFound {
		t.Errorf("UpdateWallet() error = %v, want ErrWalletNotFound", err)
	}
}

func TestWalletStore_UpdateBalance_Deposit(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	repo.CreateWallet(ctx, w)

	err := repo.UpdateBalance(ctx, "W1", 500, entity.TransactionTypeDeposit)
	if err != nil {
		t.Fatalf("UpdateBalance(Deposit) error = %v", err)
	}

	got, _ := repo.GetWalletByID(ctx, "W1")
	if got.Balance != 500 {
		t.Errorf("Balance = %d, want 500", got.Balance)
	}
	if got.Version != 2 {
		t.Errorf("Version = %d, want 2", got.Version)
	}
}

func TestWalletStore_UpdateBalance_Withdraw(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	repo.CreateWallet(ctx, w)
	repo.UpdateBalance(ctx, "W1", 1000, entity.TransactionTypeDeposit)

	err := repo.UpdateBalance(ctx, "W1", 300, entity.TransactionTypeWithdraw)
	if err != nil {
		t.Fatalf("UpdateBalance(Withdraw) error = %v", err)
	}

	got, _ := repo.GetWalletByID(ctx, "W1")
	if got.Balance != 700 {
		t.Errorf("Balance = %d, want 700", got.Balance)
	}
}

func TestWalletStore_UpdateBalance_InsufficientBalance(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	repo.CreateWallet(ctx, w)
	repo.UpdateBalance(ctx, "W1", 100, entity.TransactionTypeDeposit)

	err := repo.UpdateBalance(ctx, "W1", 200, entity.TransactionTypeWithdraw)
	if err != errors.ErrInsufficientBalance {
		t.Errorf("UpdateBalance() error = %v, want ErrInsufficientBalance", err)
	}

	// balance should remain unchanged
	got, _ := repo.GetWalletByID(ctx, "W1")
	if got.Balance != 100 {
		t.Errorf("Balance = %d, want 100 (unchanged)", got.Balance)
	}
}

func TestWalletStore_UpdateBalance_InvalidType(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	repo.CreateWallet(ctx, w)

	err := repo.UpdateBalance(ctx, "W1", 100, entity.TransactionTypeTransfer)
	if err != errors.ErrInvalidTransaction {
		t.Errorf("UpdateBalance(Transfer) error = %v, want ErrInvalidTransaction", err)
	}
}

func TestWalletStore_UpdateBalance_NotFound(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	err := repo.UpdateBalance(ctx, "W999", 100, entity.TransactionTypeDeposit)
	if err != errors.ErrWalletNotFound {
		t.Errorf("UpdateBalance() error = %v, want ErrWalletNotFound", err)
	}
}

func TestTransactionStore_CreateAndGet(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	tx := entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1")
	if err := repo.CreateTransaction(ctx, tx); err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}

	got, err := repo.GetTransaction(ctx, "T1")
	if err != nil {
		t.Fatalf("GetTransaction() error = %v", err)
	}
	if got.ID != "T1" {
		t.Errorf("ID = %q, want %q", got.ID, "T1")
	}
	if got.Amount != 100 {
		t.Errorf("Amount = %d, want 100", got.Amount)
	}
}

func TestTransactionStore_GetNotFound(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	_, err := repo.GetTransaction(ctx, "T999")
	if err != errors.ErrTransactionNotFound {
		t.Errorf("GetTransaction() error = %v, want ErrTransactionNotFound", err)
	}
}

func TestTransactionStore_IdempotencyKey(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	tx := entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1")
	repo.CreateTransaction(ctx, tx)

	// found by idempotency key
	got, err := repo.GetTransactionByIdempotencyKey(ctx, "W1", "key1")
	if err != nil {
		t.Fatalf("GetTransactionByIdempotencyKey() error = %v", err)
	}
	if got == nil || got.ID != "T1" {
		t.Errorf("expected transaction T1, got %v", got)
	}

	// not found for different key
	got, err = repo.GetTransactionByIdempotencyKey(ctx, "W1", "key-unknown")
	if err != nil {
		t.Fatalf("GetTransactionByIdempotencyKey() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for unknown key, got %v", got)
	}
}

func TestTransactionStore_ListTransactions_Pagination(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		tx := entity.NewTransaction(
			"T"+string(rune('0'+i)),
			"W1", "W2", int64(i*100),
			entity.TransactionTypeTransfer, "key"+string(rune('0'+i)),
		)
		repo.CreateTransaction(ctx, tx)
	}

	// first page
	list, total, err := repo.ListTransactions(ctx, "W1", 0, 3)
	if err != nil {
		t.Fatalf("ListTransactions() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(list) != 3 {
		t.Errorf("len = %d, want 3", len(list))
	}

	// offset beyond total
	list, total, err = repo.ListTransactions(ctx, "W1", 5, 3)
	if err != nil {
		t.Fatalf("ListTransactions() error = %v", err)
	}
	if list != nil {
		t.Errorf("expected nil for offset beyond total, got %v", list)
	}
}

func TestTransactionStore_BothWalletsIndexed(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	tx := entity.NewTransaction("T1", "W1", "W2", 100, entity.TransactionTypeTransfer, "key1")
	repo.CreateTransaction(ctx, tx)

	// should appear in W1's list
	list1, _, _ := repo.ListTransactions(ctx, "W1", 0, 10)
	if len(list1) != 1 {
		t.Errorf("W1 transactions = %d, want 1", len(list1))
	}

	// should also appear in W2's list
	list2, _, _ := repo.ListTransactions(ctx, "W2", 0, 10)
	if len(list2) != 1 {
		t.Errorf("W2 transactions = %d, want 1", len(list2))
	}
}

func TestWalletStore_Concurrent(t *testing.T) {
	repo := NewWallet()
	ctx := context.Background()

	w := entity.NewWallet("W1", "test", "hash")
	repo.CreateWallet(ctx, w)

	var wg sync.WaitGroup
	// concurrent deposits
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo.UpdateBalance(ctx, "W1", 10, entity.TransactionTypeDeposit)
		}()
	}
	// concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo.GetWalletByID(ctx, "W1")
		}()
	}
	wg.Wait()

	got, _ := repo.GetWalletByID(ctx, "W1")
	if got.Balance != 500 {
		t.Errorf("Balance = %d, want 500 after 50 deposits of 10", got.Balance)
	}
}
