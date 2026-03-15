package session

import (
	"context"
	"time"

	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/repository"
	"github.com/luojh/wallet/pkg/errors"
	"github.com/luojh/wallet/pkg/idgen"
	"github.com/luojh/wallet/pkg/pwd"
	"go.uber.org/fx"
)

type Session struct {
	sessionRepo repository.ISessionRepository
	idFactory   *idgen.IDFactory
	ttl         time.Duration
}

func NewSessionDomain(lc fx.Lifecycle, sessionRepo repository.ISessionRepository, idFactory *idgen.IDFactory, cfg *config.Config) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.StopHook(cancel))

	s := &Session{
		sessionRepo: sessionRepo,
		idFactory:   idFactory,
		ttl:         cfg.Session.TTL,
	}

	s.cleanup(ctx)

	return s
}

func (s *Session) CreateSession(ctx context.Context, walletID string, ttl ...time.Duration) (*entity.Session, error) {
	// 生成会话ID
	sessionID, err := s.idFactory.GenSessionID()
	if err != nil {
		return nil, err
	}
	ttld := s.ttl
	if len(ttl) > 0 {
		ttld = ttl[0]
	}

	session, err := entity.NewSession(sessionID, walletID, ttld)
	if err != nil {
		return nil, err
	}

	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Session) InvalidateSession(ctx context.Context, token string) error {
	return s.sessionRepo.Invalidate(ctx, token)
}

func (s *Session) ValidateSession(ctx context.Context, token string) (*entity.Session, error) {
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !session.IsValidSession() {
		return nil, errors.ErrSessionExpired
	}

	return session, nil
}

func (s *Session) GetWalletID(ctx context.Context, token string) (string, error) {
	session, err := s.ValidateSession(ctx, token)
	if err != nil {
		return "", err
	}
	return session.WalletID, nil
}

func (s *Session) VerifyPassword(ctx context.Context, passwordHash, password string) error {
	if pwd.CheckPasswordHash(password, passwordHash) {
		return nil
	}
	return errors.ErrInvalidPassword
}

func (s *Session) cleanup(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(s.ttl):
			case <-ctx.Done():
				s.sessionRepo.Cleanup(ctx)
				return
			}
		}
	}()
}
