package service

import (
	"context"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/domain/session"
	"github.com/luojh/wallet/internal/repository"
	"github.com/luojh/wallet/pkg/errors"
	"github.com/luojh/wallet/pkg/idgen"
)

type AuthService struct {
	session    *session.Session
	walletRepo repository.IWalletRepository
	idFactory  *idgen.IDFactory
}

func NewAuthService(session *session.Session, walletRepo repository.IWalletRepository, idFactory *idgen.IDFactory) IAuthService {
	return &AuthService{session: session, walletRepo: walletRepo, idFactory: idFactory}
}

// Login 登录
func (s *AuthService) Login(ctx context.Context, walletID, password string) (*entity.Session, error) {
	// 获取钱包
	wallet, err := s.walletRepo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// 验证密码
	if err := s.session.VerifyPassword(ctx, wallet.PasswordHash, password); err != nil {
		return nil, err
	}

	// 检查钱包状态
	if !wallet.CanTransfer() {
		return nil, errors.ErrWalletFrozen
	}

	// 创建会话
	session, err := s.session.CreateSession(ctx, walletID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Logout 登出
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.session.InvalidateSession(ctx, token)
}

// ValidateSession 验证会话
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*entity.Session, error) {
	return s.session.ValidateSession(ctx, token)
}

// GetWalletIDByToken 通过令牌获取钱包ID
func (s *AuthService) GetWalletID(ctx context.Context, token string) (string, error) {
	return s.session.GetWalletID(ctx, token)
}
