package entity

import (
	"time"
)

// WalletStatus 钱包状态
type WalletStatus int

const (
	WalletStatusActive WalletStatus = iota + 1
	WalletStatusFrozen
	WalletStatusClosed
)

func (s WalletStatus) String() string {
	switch s {
	case WalletStatusActive:
		return "active"
	case WalletStatusFrozen:
		return "frozen"
	case WalletStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Wallet 钱包实体
type Wallet struct {
	ID           string       // 钱包唯一标识（雪花ID）
	Name         string       // 钱包名称
	PasswordHash string       // 密码哈希（bcrypt）
	Balance      int64        // 余额（单位：分，避免浮点精度问题）
	Status       WalletStatus // 状态
	CreatedAt    time.Time    // 创建时间
	UpdatedAt    time.Time    // 更新时间
	Version      int64        // 乐观锁版本号
}

// NewWallet 创建新钱包
func NewWallet(id, name, passwordHash string) *Wallet {
	now := time.Now()
	return &Wallet{
		ID:           id,
		Name:         name,
		PasswordHash: passwordHash,
		Balance:      0,
		Status:       WalletStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
}

// CanTransfer 检查钱包是否可以转账
func (w *Wallet) CanTransfer() bool {
	return w.Status == WalletStatusActive
}

// HasSufficientBalance 检查余额是否充足
func (w *Wallet) HasSufficientBalance(amount int64) bool {
	return w.Balance >= amount
}
