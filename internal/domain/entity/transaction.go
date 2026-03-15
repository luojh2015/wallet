package entity

import (
	"time"
)

// TransactionType 交易类型
type TransactionType int

const (
	TransactionTypeTransfer TransactionType = iota + 1
	TransactionTypeDeposit
	TransactionTypeWithdraw
)

func (t TransactionType) String() string {
	switch t {
	case TransactionTypeTransfer:
		return "transfer"
	case TransactionTypeDeposit:
		return "deposit"
	case TransactionTypeWithdraw:
		return "withdraw"
	default:
		return "unknown"
	}
}

// TransactionStatus 交易状态
type TransactionStatus int

const (
	TransactionStatusPending TransactionStatus = iota + 1
	TransactionStatusCompleted
	TransactionStatusFailed
	TransactionStatusCancelled
)

func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPending:
		return "pending"
	case TransactionStatusCompleted:
		return "completed"
	case TransactionStatusFailed:
		return "failed"
	case TransactionStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Transaction 交易记录实体
type Transaction struct {
	ID             string            // 交易唯一标识（雪花ID）
	FromWalletID   string            // 源钱包ID
	ToWalletID     string            // 目标钱包ID
	Amount         int64             // 交易金额（分）
	Type           TransactionType   // 类型
	Status         TransactionStatus // 状态
	IdempotencyKey string            // 幂等键
	CreatedAt      time.Time         // 创建时间
	CompletedAt    *time.Time        // 完成时间
}

// NewTransaction 创建新交易记录
func NewTransaction(id, fromWalletID, toWalletID string, amount int64, typ TransactionType, idempotencyKey string) *Transaction {
	return &Transaction{
		ID:             id,
		FromWalletID:   fromWalletID,
		ToWalletID:     toWalletID,
		Amount:         amount,
		Type:           typ,
		Status:         TransactionStatusPending,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
}

// MarkCompleted 标记交易完成
func (t *Transaction) MarkCompleted() {
	now := time.Now()
	t.Status = TransactionStatusCompleted
	t.CompletedAt = &now
}

// MarkFailed 标记交易失败
func (t *Transaction) MarkFailed() {
	now := time.Now()
	t.Status = TransactionStatusFailed
	t.CompletedAt = &now
}

// MarkCancelled 标记交易取消
func (t *Transaction) MarkCancelled() {
	now := time.Now()
	t.Status = TransactionStatusCancelled
	t.CompletedAt = &now
}
