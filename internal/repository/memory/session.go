package memory

import (
	"context"
	"sync"
	"time"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/internal/repository"
	"github.com/luojh/wallet/pkg/errors"
)

type session struct {
	mu       sync.RWMutex
	sessions map[string]*entity.Session // token -> session
	byID     map[string]*entity.Session // id -> session
}

func NewSession() repository.ISessionRepository {
	return &session{
		sessions: make(map[string]*entity.Session),
		byID:     make(map[string]*entity.Session),
	}
}

func (r *session) Create(ctx context.Context, session *entity.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := *session
	r.sessions[session.Token] = &s
	r.byID[session.ID] = &s

	return nil
}

func (r *session) GetByToken(ctx context.Context, token string) (*entity.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[token]
	if !exists {
		return nil, errors.ErrInvalidToken
	}

	s := *session
	return &s, nil
}

func (r *session) GetByWalletID(ctx context.Context, id string) (*entity.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.byID[id]
	if !exists {
		return nil, errors.ErrInvalidToken
	}

	s := *session
	return &s, nil
}

func (r *session) Invalidate(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[token]
	if !exists {
		return nil // 已不存在，视为成功
	}

	session.IsValid = false

	return nil
}

func (r *session) InvalidateByWalletID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, session := range r.sessions {
		if session.WalletID == id {
			session.IsValid = false
		}
	}

	return nil

}

func (r *session) Cleanup(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for token, session := range r.sessions {
		if session.ExpiresAt.Before(now) || !session.IsValid {
			delete(r.sessions, token)
			delete(r.byID, session.ID)
		}
	}

	return nil
}
