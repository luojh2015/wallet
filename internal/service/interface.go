package service

import (
	"context"

	"github.com/luojh/wallet/internal/domain/entity"
)

type IAuthService interface {
	Login(ctx context.Context, walletID string, password string) (*entity.Session, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*entity.Session, error)
	GetWalletID(ctx context.Context, token string) (string, error)
}

type IWalletService interface {
	CreateWallet(ctx context.Context, name string, password string) (string, error)
	GetWallet(ctx context.Context, id string) (*entity.Wallet, error)
	Deposit(ctx context.Context, walletID, idempotencyKey string, amount float64) error
	Withdraw(ctx context.Context, walletID, idempotencyKey string, amount float64) error
	Transfer(ctx context.Context, fromWalletID, toWalletID, idempotencyKey string, amount float64) error
	GetTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, error)
}
