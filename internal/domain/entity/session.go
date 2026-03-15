package entity

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Session 会话实体
type Session struct {
	ID        string    // 会话ID
	WalletID  string    // 关联钱包ID
	Token     string    // 会话令牌
	ExpiresAt time.Time // 过期时间
	CreatedAt time.Time // 创建时间
	IsValid   bool      // 是否有效
}

// NewSession 创建新会话
func NewSession(id, walletID string, ttl time.Duration) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Session{
		ID:        id,
		WalletID:  walletID,
		Token:     token,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
		IsValid:   true,
	}, nil
}

// IsExpired 检查会话是否过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValidSession 检查会话是否有效
func (s *Session) IsValidSession() bool {
	return s.IsValid && !s.IsExpired()
}

// Invalidate 使会话失效
func (s *Session) Invalidate() {
	s.IsValid = false
}

// generateToken 生成随机令牌
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
