package wallet

import (
	"context"
	"sync"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/repository"
	"github.com/luojh/wallet/pkg/errors"
	"github.com/luojh/wallet/pkg/idgen"
	"github.com/luojh/wallet/pkg/pwd"
)

type Wallet struct {
	walletRepo repository.IWalletRepository
	idFactory  *idgen.IDFactory
	sync.Mutex
}

func NewWallet(walletRepo repository.IWalletRepository, idFactory *idgen.IDFactory) *Wallet {
	return &Wallet{walletRepo: walletRepo, idFactory: idFactory}
}

func (w *Wallet) CreateWallet(ctx context.Context, name, password string) (*entity.Wallet, error) {
	id, err := w.idFactory.GenWalletID()
	if err != nil {
		return nil, err
	}
	pwdHash, err := pwd.HashPassword(password)
	if err != nil {
		return nil, err
	}
	wallet := entity.NewWallet(id, name, pwdHash)

	if err = w.walletRepo.CreateWallet(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (w *Wallet) GetWallet(ctx context.Context, id string) (*entity.Wallet, error) {
	return w.walletRepo.GetWalletByID(ctx, id)
}

func (w *Wallet) UpdateWallet(ctx context.Context, id string, name, pass string) error {
	if pass != "" {
		var err error
		pass, err = pwd.HashPassword(pass)
		if err != nil {
			return err
		}
	}
	return w.walletRepo.UpdateWallet(ctx, id, name, pass)
}

// 存款
func (w *Wallet) Deposit(ctx context.Context, wallet_id, idempotencyKey string, amount int64) (*entity.Transaction, error) {
	w.Lock()
	defer w.Unlock()

	ts, err := w.walletRepo.GetTransactionByIdempotencyKey(ctx, wallet_id, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if ts != nil {
		return ts, nil
	}

	id, err := w.idFactory.GenTransactionID()
	ts = entity.NewTransaction(id, wallet_id, "", amount, entity.TransactionTypeTransfer, idempotencyKey)

	if err = w.walletRepo.UpdateBalance(ctx, wallet_id, amount, entity.TransactionTypeDeposit); err != nil {
		ts.MarkFailed()
	} else {
		ts.MarkCompleted()
	}
	err = w.walletRepo.CreateTransaction(ctx, ts)
	if err != nil {
		return nil, err
	}
	return ts, nil
}

// 取款
func (w *Wallet) Withdraw(ctx context.Context, wallet_id, idempotencyKey string, amount int64) (*entity.Transaction, error) {
	w.Lock()
	defer w.Unlock()

	ts, err := w.walletRepo.GetTransactionByIdempotencyKey(ctx, wallet_id, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if ts != nil {
		return ts, nil
	}

	id, err := w.idFactory.GenTransactionID()
	ts = entity.NewTransaction(id, wallet_id, "", amount, entity.TransactionTypeWithdraw, idempotencyKey)

	if err = w.walletRepo.UpdateBalance(ctx, wallet_id, amount, entity.TransactionTypeWithdraw); err != nil {
		ts.MarkFailed()
	} else {
		ts.MarkCompleted()
	}
	err = w.walletRepo.CreateTransaction(ctx, ts)
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func (w *Wallet) Transfer(ctx context.Context, fromWalletID, toWalletID, idempotencyKey string, amount int64) (*entity.Transaction, error) {
	w.Lock()
	defer w.Unlock()

	ts, err := w.walletRepo.GetTransactionByIdempotencyKey(ctx, fromWalletID, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if ts != nil {
		return ts, nil
	}

	id, err := w.idFactory.GenTransactionID()
	ts = entity.NewTransaction(id, fromWalletID, toWalletID, amount, entity.TransactionTypeTransfer, idempotencyKey)

	if err := w.executeTransfer(ctx, ts, amount); err != nil {
		ts.MarkFailed()
	} else {
		ts.MarkCompleted()
	}
	err = w.walletRepo.CreateTransaction(ctx, ts)
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func (w *Wallet) executeTransfer(ctx context.Context, tx *entity.Transaction, amount int64) error {
	fromWallet, err := w.walletRepo.GetWalletByID(ctx, tx.FromWalletID)
	if err != nil {
		return err
	}

	// 再次检查余额（在锁内）
	if !fromWallet.HasSufficientBalance(amount) {
		return errors.ErrInsufficientBalance
	}

	// 扣减源账户余额
	if err := w.walletRepo.UpdateBalance(ctx, tx.FromWalletID, amount, entity.TransactionTypeWithdraw); err != nil {
		return err
	}

	// 增加目标账户余额
	// 回滚源账户
	if err := w.walletRepo.UpdateBalance(ctx, tx.ToWalletID, amount, entity.TransactionTypeDeposit); err != nil {
		_ = w.walletRepo.UpdateBalance(ctx, tx.FromWalletID, amount, entity.TransactionTypeDeposit)
		return err
	}

	return nil
}

func (w *Wallet) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	return w.walletRepo.GetTransaction(ctx, id)
}

func (w *Wallet) ListTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
	return w.walletRepo.ListTransactions(ctx, walletID, offset, limit)
}
