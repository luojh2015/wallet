package memory

import (
	"context"
	"sync"
	"time"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/repository"
	"github.com/luojh/wallet/pkg/errors"
)

type wallet struct {
	ws *walletStore
	ts *transactionStore
}

func NewWallet() repository.IWalletRepository {
	return &wallet{
		ws: NewWalletStore(),
		ts: NewTransactionStore(),
	}
}

func (r *wallet) CreateWallet(ctx context.Context, wallet *entity.Wallet) error {
	return r.ws.CreateWallet(ctx, wallet)
}

func (r *wallet) GetWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	return r.ws.GetWalletByID(ctx, id)
}

func (r *wallet) UpdateWallet(ctx context.Context, id string, name, hashPwd string) error {
	return r.ws.UpdateWallet(ctx, id, name, hashPwd)
}

func (r *wallet) UpdateBalance(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
	return r.ws.UpdateBalance(ctx, id, amount, typ)
}

func (r *wallet) GetTransactionByIdempotencyKey(ctx context.Context, walletID, key string) (*entity.Transaction, error) {
	return r.ts.GetTransactionByIdempotencyKey(ctx, walletID, key)
}

func (r *wallet) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	return r.ts.GetTransaction(ctx, id)
}

func (r *wallet) CreateTransaction(ctx context.Context, ts *entity.Transaction) error {
	return r.ts.CreateTransaction(ctx, ts)
}

func (r *wallet) ListTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
	return r.ts.ListTransactions(ctx, walletID, offset, limit)
}

type walletStore struct {
	mu      sync.RWMutex
	wallets map[string]*entity.Wallet
}

func NewWalletStore() *walletStore {
	return &walletStore{
		mu:      sync.RWMutex{},
		wallets: make(map[string]*entity.Wallet),
	}
}

func (r *walletStore) CreateWallet(ctx context.Context, wallet *entity.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.wallets[wallet.ID]; exists {
		return errors.ErrWalletExists
	}

	w := *wallet
	r.wallets[wallet.ID] = &w

	return nil
}

func (r *walletStore) GetWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wallet, exists := r.wallets[id]
	if !exists {
		return nil, errors.ErrWalletNotFound
	}

	// 返回拷贝
	w := *wallet
	return &w, nil
}

func (r *walletStore) UpdateWallet(ctx context.Context, id string, name, hashPwd string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wallet, exists := r.wallets[id]
	if !exists {
		return errors.ErrWalletNotFound
	}
	wallet.Name = name
	wallet.PasswordHash = hashPwd
	wallet.Version++
	wallet.UpdatedAt = time.Now()

	return nil
}

func (r *walletStore) UpdateBalance(ctx context.Context, id string, amount int64, typ entity.TransactionType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wallet, exists := r.wallets[id]
	if !exists {
		return errors.ErrWalletNotFound
	}

	switch typ {
	case entity.TransactionTypeDeposit:
		wallet.Balance += amount
	case entity.TransactionTypeWithdraw:
		if wallet.Balance < amount {
			return errors.ErrInsufficientBalance
		}
		wallet.Balance -= amount
	default:
		return errors.ErrInvalidTransaction
	}

	wallet.Version++
	wallet.UpdatedAt = time.Now()

	return nil
}

type transactionStore struct {
	mu               sync.RWMutex
	transactions     map[string]*entity.Transaction
	byWallet         map[string][]*entity.Transaction
	byIdempotencyKey map[string]*entity.Transaction
}

func NewTransactionStore() *transactionStore {
	return &transactionStore{
		mu:               sync.RWMutex{},
		transactions:     make(map[string]*entity.Transaction),
		byWallet:         make(map[string][]*entity.Transaction),
		byIdempotencyKey: make(map[string]*entity.Transaction),
	}
}

func (r *transactionStore) GetTransactionByIdempotencyKey(ctx context.Context, walletID, key string) (*entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.byIdempotencyKey[walletID+":"+key], nil
}

func (r *transactionStore) GetTransaction(ctx context.Context, id string) (*entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.transactions[id]
	if !exists {
		return nil, errors.ErrTransactionNotFound
	}

	n := *t

	return &n, nil
}

func (r *transactionStore) CreateTransaction(ctx context.Context, ts *entity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[ts.ID] = ts
	if ts.FromWalletID != "" {
		r.byWallet[ts.FromWalletID] = append(r.byWallet[ts.FromWalletID], ts)
	}

	if ts.ToWalletID != "" && ts.FromWalletID != ts.ToWalletID {
		r.byWallet[ts.ToWalletID] = append(r.byWallet[ts.ToWalletID], ts)
	}
	r.byIdempotencyKey[ts.FromWalletID+":"+ts.IdempotencyKey] = ts

	return nil
}

func (r *transactionStore) ListTransactions(ctx context.Context, walletID string, offset, limit int) ([]*entity.Transaction, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transactions := r.byWallet[walletID]
	total := len(transactions)
	if total <= offset {
		return nil, 0, nil
	}

	if total < offset+limit {
		limit = total - offset
	}

	res := make([]*entity.Transaction, limit)

	// 倒序
	for i := offset; i < offset+limit; i++ {
		t := *transactions[i]
		res[limit-i-1] = &t
	}

	return res, int64(total), nil
}
