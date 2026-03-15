package scripts

import (
	"context"
	"testing"
	"time"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/pkg/errors"
)

func TestFullWalletLifecycle(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	// 1. 创建钱包
	w := mustCreateWallet(t, env, "Alice", "pass123")
	if w.Name != "Alice" {
		t.Errorf("Name = %q, want %q", w.Name, "Alice")
	}
	if w.Balance != 0 {
		t.Errorf("initial balance = %d, want 0", w.Balance)
	}
	if w.Status != entity.WalletStatusActive {
		t.Errorf("Status = %d, want Active(%d)", w.Status, entity.WalletStatusActive)
	}

	// 2. 存款
	tx1 := mustDeposit(t, env, w.ID, "deposit-1", 10000)
	if tx1.Status != entity.TransactionStatusCompleted {
		t.Errorf("deposit status = %d, want Completed(%d)", tx1.Status, entity.TransactionStatusCompleted)
	}
	if tx1.Amount != 10000 {
		t.Errorf("deposit amount = %d, want 10000", tx1.Amount)
	}
	assertBalance(t, env, w.ID, 10000)

	// 3. 取款
	tx2, err := env.walletService.Withdraw(ctx, w.ID, "withdraw-1", 3000)
	if err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
	if tx2.Status != entity.TransactionStatusCompleted {
		t.Errorf("withdraw status = %d, want Completed(%d)", tx2.Status, entity.TransactionStatusCompleted)
	}
	assertBalance(t, env, w.ID, 7000)

	// 4. 查交易列表
	txs, total, err := env.walletService.ListTransactions(ctx, w.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if total != 2 {
		t.Errorf("total transactions = %d, want 2", total)
	}
	if len(txs) != 2 {
		t.Errorf("returned transactions = %d, want 2", len(txs))
	}
}

func TestLoginTransferLogout(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	// 创建两个钱包
	alice := mustCreateWallet(t, env, "Alice", "passA")
	bob := mustCreateWallet(t, env, "Bob", "passB")
	mustDeposit(t, env, alice.ID, "init-deposit", 5000)

	// 登录
	sess, err := env.authService.Login(ctx, alice.ID, "passA")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if sess == nil || sess.Token == "" {
		t.Fatal("Login returned nil session or empty token")
	}

	// 转账
	tx, err := env.walletService.Transfer(ctx, alice.ID, bob.ID, "transfer-1", 2000)
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if tx.Status != entity.TransactionStatusCompleted {
		t.Errorf("transfer status = %d, want Completed", tx.Status)
	}

	// 验证余额
	assertBalance(t, env, alice.ID, 3000)
	assertBalance(t, env, bob.ID, 2000)

	// 查交易
	txsA, totalA, err := env.walletService.ListTransactions(ctx, alice.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListTransactions(A): %v", err)
	}
	if totalA < 2 {
		t.Errorf("Alice total transactions = %d, want >= 2", totalA)
	}
	_ = txsA

	txsB, totalB, err := env.walletService.ListTransactions(ctx, bob.ID, 0, 10)
	if err != nil {
		t.Fatalf("ListTransactions(B): %v", err)
	}
	if totalB < 1 {
		t.Errorf("Bob total transactions = %d, want >= 1", totalB)
	}
	_ = txsB

	// 登出
	if err := env.authService.Logout(ctx, sess.Token); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// 验证会话失效
	_, err = env.authService.ValidateSession(ctx, sess.Token)
	if err == nil {
		t.Error("ValidateSession after logout should return error")
	}
}

func TestDepositIdempotency(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	w := mustCreateWallet(t, env, "Alice", "pass")

	tx1, err := env.walletService.Deposit(ctx, w.ID, "deposit-idem-1", 500)
	if err != nil {
		t.Fatalf("first Deposit: %v", err)
	}

	tx2, err := env.walletService.Deposit(ctx, w.ID, "deposit-idem-1", 500)
	if err != nil {
		t.Fatalf("second Deposit: %v", err)
	}

	if tx1.ID != tx2.ID {
		t.Errorf("idempotent deposit returned different tx IDs: %q vs %q", tx1.ID, tx2.ID)
	}

	assertBalance(t, env, w.ID, 500)
}

func TestWithdrawIdempotency(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	w := mustCreateWallet(t, env, "Alice", "pass")
	mustDeposit(t, env, w.ID, "init", 1000)

	tx1, err := env.walletService.Withdraw(ctx, w.ID, "withdraw-idem-1", 300)
	if err != nil {
		t.Fatalf("first Withdraw: %v", err)
	}

	tx2, err := env.walletService.Withdraw(ctx, w.ID, "withdraw-idem-1", 300)
	if err != nil {
		t.Fatalf("second Withdraw: %v", err)
	}

	if tx1.ID != tx2.ID {
		t.Errorf("idempotent withdraw returned different tx IDs: %q vs %q", tx1.ID, tx2.ID)
	}

	assertBalance(t, env, w.ID, 700)
}

func TestTransferIdempotency(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	a := mustCreateWallet(t, env, "Alice", "pass")
	b := mustCreateWallet(t, env, "Bob", "pass")
	mustDeposit(t, env, a.ID, "init", 2000)

	tx1, err := env.walletService.Transfer(ctx, a.ID, b.ID, "transfer-idem-1", 800)
	if err != nil {
		t.Fatalf("first Transfer: %v", err)
	}

	tx2, err := env.walletService.Transfer(ctx, a.ID, b.ID, "transfer-idem-1", 800)
	if err != nil {
		t.Fatalf("second Transfer: %v", err)
	}

	if tx1.ID != tx2.ID {
		t.Errorf("idempotent transfer returned different tx IDs: %q vs %q", tx1.ID, tx2.ID)
	}

	assertBalance(t, env, a.ID, 1200)
	assertBalance(t, env, b.ID, 800)
}

func TestTransferInsufficientBalance(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	a := mustCreateWallet(t, env, "Alice", "pass")
	b := mustCreateWallet(t, env, "Bob", "pass")
	mustDeposit(t, env, a.ID, "init", 500)

	tx, err := env.walletService.Transfer(ctx, a.ID, b.ID, "over-transfer", 1000)
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if tx.Status != entity.TransactionStatusFailed {
		t.Errorf("transfer status = %d, want Failed(%d)", tx.Status, entity.TransactionStatusFailed)
	}

	assertBalance(t, env, a.ID, 500)
	assertBalance(t, env, b.ID, 0)
}

func TestTransferTargetNotFound(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	a := mustCreateWallet(t, env, "Alice", "pass")
	mustDeposit(t, env, a.ID, "init", 1000)

	tx, err := env.walletService.Transfer(ctx, a.ID, "nonexistent-wallet", "bad-target", 500)
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if tx.Status != entity.TransactionStatusFailed {
		t.Errorf("transfer status = %d, want Failed(%d)", tx.Status, entity.TransactionStatusFailed)
	}

	// 源钱包余额应回滚
	assertBalance(t, env, a.ID, 1000)
}

func TestSessionExpiration(t *testing.T) {
	env := newTestEnvWithSessionTTL(t, 50*time.Millisecond)
	ctx := context.Background()

	w := mustCreateWallet(t, env, "Alice", "pass")

	sess, err := env.authService.Login(ctx, w.ID, "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// 验证会话当前有效
	_, err = env.authService.ValidateSession(ctx, sess.Token)
	if err != nil {
		t.Fatalf("ValidateSession before expiry: %v", err)
	}

	// 等待过期
	time.Sleep(100 * time.Millisecond)

	_, err = env.authService.ValidateSession(ctx, sess.Token)
	if !errors.IsAppError(err) {
		t.Errorf("ValidateSession after expiry: got %v, want AppError", err)
	}
}

func TestSessionInvalidation(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	w := mustCreateWallet(t, env, "Alice", "pass")
	sess, err := env.authService.Login(ctx, w.ID, "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// 验证有效
	_, err = env.authService.ValidateSession(ctx, sess.Token)
	if err != nil {
		t.Fatalf("ValidateSession before logout: %v", err)
	}

	// 登出
	if err := env.authService.Logout(ctx, sess.Token); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// 验证失效
	_, err = env.authService.ValidateSession(ctx, sess.Token)
	if err == nil {
		t.Error("ValidateSession after logout should return error")
	}
}

func TestMultiWalletChainTransfer(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	a := mustCreateWallet(t, env, "Alice", "pass")
	b := mustCreateWallet(t, env, "Bob", "pass")
	c := mustCreateWallet(t, env, "Charlie", "pass")

	mustDeposit(t, env, a.ID, "init", 10000)

	// A -> B: 4000
	tx1, err := env.walletService.Transfer(ctx, a.ID, b.ID, "a-to-b", 4000)
	if err != nil {
		t.Fatalf("Transfer A->B: %v", err)
	}
	if tx1.Status != entity.TransactionStatusCompleted {
		t.Fatalf("A->B status = %d, want Completed", tx1.Status)
	}

	// B -> C: 2000
	tx2, err := env.walletService.Transfer(ctx, b.ID, c.ID, "b-to-c", 2000)
	if err != nil {
		t.Fatalf("Transfer B->C: %v", err)
	}
	if tx2.Status != entity.TransactionStatusCompleted {
		t.Fatalf("B->C status = %d, want Completed", tx2.Status)
	}

	// 验证余额
	assertBalance(t, env, a.ID, 6000)
	assertBalance(t, env, b.ID, 2000)
	assertBalance(t, env, c.ID, 2000)

	// 总余额守恒
	balA := getBalance(t, env, a.ID)
	balB := getBalance(t, env, b.ID)
	balC := getBalance(t, env, c.ID)
	total := balA + balB + balC
	if total != 10000 {
		t.Errorf("total balance = %d, want 10000 (A=%d, B=%d, C=%d)", total, balA, balB, balC)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	w := mustCreateWallet(t, env, "Alice", "correct")

	_, err := env.authService.Login(ctx, w.ID, "wrong")
	if err == nil {
		t.Fatal("Login with wrong password should return error")
	}
	if err != errors.ErrInvalidPassword {
		t.Errorf("Login error = %v, want ErrInvalidPassword", err)
	}
}

func TestLoginFrozenWallet(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	w := mustCreateFrozenWallet(t, env, "Frozen", "pass")

	_, err := env.authService.Login(ctx, w.ID, "pass")
	if err == nil {
		t.Fatal("Login to frozen wallet should return error")
	}
	if err != errors.ErrWalletFrozen {
		t.Errorf("Login error = %v, want ErrWalletFrozen", err)
	}
}
