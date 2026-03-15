package service

import (
	"context"
	"errors"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/domain/wallet"
)

type WalletService struct {
	wallet *wallet.Wallet
}

func NewWalletService(wallet *wallet.Wallet) *WalletService {
	return &WalletService{wallet: wallet}
}

func (s *WalletService) CreateWallet(ctx context.Context, name string, password string) (*entity.Wallet, error) {
	return s.wallet.CreateWallet(ctx, name, password)
}

func (s *WalletService) GetWallet(ctx context.Context, id string) (*entity.Wallet, error) {
	return s.wallet.GetWallet(ctx, id)
}

func (s *WalletService) UpdateWallet(ctx context.Context, id string, name string, password string) error {
	return s.wallet.UpdateWallet(ctx, id, name, password)
}

func (s *WalletService) Deposit(ctx context.Context, id, idempotencyKey string, amount int64) (*entity.Transaction, error) {
	if id == "" || idempotencyKey == "" || amount <= 0 {
		return nil, errors.New("invalid parameter")
	}

	return s.wallet.Deposit(ctx, id, idempotencyKey, amount)
}

func (s *WalletService) Withdraw(ctx context.Context, id, idempotencyKey string, amount int64) (*entity.Transaction, error) {
	if id == "" || idempotencyKey == "" || amount <= 0 {
		return nil, errors.New("invalid parameter")
	}

	return s.wallet.Withdraw(ctx, id, idempotencyKey, amount)
}

func (s *WalletService) Transfer(ctx context.Context, fromWalletID, toWalletID, idempotencyKey string, amount int64) (*entity.Transaction, error) {
	if fromWalletID == "" || toWalletID == "" || fromWalletID == toWalletID || idempotencyKey == "" || amount <= 0 {
		return nil, errors.New("invalid parameter")
	}
	return s.wallet.Transfer(ctx, fromWalletID, toWalletID, idempotencyKey, amount)
}

func (s *WalletService) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	return s.wallet.GetTransaction(ctx, id)
}

func (s *WalletService) ListTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
	return s.wallet.ListTransactions(ctx, walletID, offset, limit)
}
