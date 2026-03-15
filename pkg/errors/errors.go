package errors

import (
	"errors"
	"fmt"
)

// AppError 应用错误
type AppError struct {
	Code    int    // 错误码
	Message string // 错误消息
	Detail  string // 详细信息
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewAppError 创建应用错误
func NewAppError(code int, message, detail string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// 错误码定义
const (
	// 1xxx - 通用错误
	CodeInternalError      = 1001
	CodeInvalidRequest     = 1002
	CodeServiceUnavailable = 1003
	CodeNotImplemented     = 1004

	// 2xxx - 钱包相关错误
	CodeWalletNotFound  = 2001
	CodeWalletExists    = 2002
	CodeWalletFrozen    = 2003
	CodeInvalidPassword = 2004

	// 3xxx - 交易相关错误
	CodeInsufficientBalance  = 3001
	CodeInvalidAmount        = 3002
	CodeTargetWalletNotFound = 3003
	CodeTransactionNotFound  = 3004
	CodeDuplicateTransaction = 3005
	CodeInvalidTransaction   = 3006

	// 4xxx - 认证相关错误
	CodeUnauthorized   = 4001
	CodeSessionExpired = 4002
	CodeInvalidToken   = 4003

	// 5xxx - 并发相关错误
	CodeLockTimeout    = 5001
	CodeOptimisticLock = 5002
	CodeProcessing     = 5003
)

// 预定义错误
var (
	ErrInternalError      = NewAppError(CodeInternalError, "内部服务器错误", "")
	ErrInvalidRequest     = NewAppError(CodeInvalidRequest, "请求参数无效", "")
	ErrServiceUnavailable = NewAppError(CodeServiceUnavailable, "服务不可用", "")
	ErrNotImplemented     = NewAppError(CodeNotImplemented, "功能未实现", "")

	ErrWalletNotFound  = NewAppError(CodeWalletNotFound, "钱包不存在", "")
	ErrWalletExists    = NewAppError(CodeWalletExists, "钱包已存在", "")
	ErrWalletFrozen    = NewAppError(CodeWalletFrozen, "钱包已冻结", "")
	ErrInvalidPassword = NewAppError(CodeInvalidPassword, "密码错误", "")

	ErrInsufficientBalance  = NewAppError(CodeInsufficientBalance, "余额不足", "")
	ErrInvalidAmount        = NewAppError(CodeInvalidAmount, "转账金额无效", "")
	ErrTargetWalletNotFound = NewAppError(CodeTargetWalletNotFound, "目标钱包不存在", "")
	ErrTransactionNotFound  = NewAppError(CodeTransactionNotFound, "交易不存在", "")
	ErrDuplicateTransaction = NewAppError(CodeDuplicateTransaction, "重复的交易请求", "")
	ErrInvalidTransaction   = NewAppError(CodeInvalidTransaction, "无效的交易", "")

	ErrUnauthorized   = NewAppError(CodeUnauthorized, "未认证", "")
	ErrSessionExpired = NewAppError(CodeSessionExpired, "会话过期", "")
	ErrInvalidToken   = NewAppError(CodeInvalidToken, "无效令牌", "")

	ErrLockTimeout    = NewAppError(CodeLockTimeout, "获取锁超时", "")
	ErrOptimisticLock = NewAppError(CodeOptimisticLock, "数据已被修改，请重试", "")
	ErrProcessing     = NewAppError(CodeProcessing, "请求处理中", "")
)

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError 获取应用错误
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternalError
}

// WithDetail 添加详细信息
func WithDetail(err *AppError, detail string) *AppError {
	return &AppError{
		Code:    err.Code,
		Message: err.Message,
		Detail:  detail,
	}
}
