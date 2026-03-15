package repository

import (
	"context"

	"github.com/luojh/wallet/internal/domain/entity"
)

type ISessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error

	GetByToken(ctx context.Context, token string) (*entity.Session, error)

	GetByWalletID(ctx context.Context, id string) (*entity.Session, error)

	Invalidate(ctx context.Context, token string) error

	InvalidateByWalletID(ctx context.Context, id string) error

	Cleanup(ctx context.Context) error
}

type IWalletRepository interface {
	CreateWallet(ctx context.Context, wallet *entity.Wallet) error

	GetWalletByID(ctx context.Context, id string) (*entity.Wallet, error)

	UpdateWallet(ctx context.Context, id string, name, hashPwd string) error

	UpdateBalance(ctx context.Context, id string, amount int64, typ entity.TransactionType) error

	GetTransactionByIdempotencyKey(ctx context.Context, walletID, key string) (*entity.Transaction, error)

	GetTransaction(ctx context.Context, id string) (*entity.Transaction, error)

	CreateTransaction(ctx context.Context, ts *entity.Transaction) error

	ListTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error)
}
