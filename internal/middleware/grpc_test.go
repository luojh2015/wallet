package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/luojh/wallet/internal/domain/entity"
	"github.com/luojh/wallet/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// --- mock AuthService ---

type mockAuthService struct {
	loginFn           func(ctx context.Context, walletID, password string) (*entity.Session, error)
	logoutFn          func(ctx context.Context, token string) error
	validateSessionFn func(ctx context.Context, token string) (*entity.Session, error)
	getWalletIDFn     func(ctx context.Context, token string) (string, error)
}

func (m *mockAuthService) Login(ctx context.Context, walletID, password string) (*entity.Session, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, walletID, password)
	}
	return nil, nil
}
func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, token)
	}
	return nil
}
func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*entity.Session, error) {
	if m.validateSessionFn != nil {
		return m.validateSessionFn(ctx, token)
	}
	return nil, nil
}
func (m *mockAuthService) GetWalletID(ctx context.Context, token string) (string, error) {
	if m.getWalletIDFn != nil {
		return m.getWalletIDFn(ctx, token)
	}
	return "", nil
}

// --- Tests for UnaryInterceptor ---

func TestGRPCAuthInterceptor_UnauthenticatedMethod(t *testing.T) {
	auth := &mockAuthService{}
	interceptor := NewGRPCAuthInterceptor(auth)

	var handlerCalled bool
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/wallet.v1.WalletService/CreateWallet"}
	resp, err := interceptor.UnaryInterceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("UnaryInterceptor() error = %v", err)
	}
	if !handlerCalled {
		t.Error("handler should be called for unauthenticated method")
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want 'ok'", resp)
	}
}

func TestGRPCAuthInterceptor_AuthenticatedMethod_Success(t *testing.T) {
	auth := &mockAuthService{
		validateSessionFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return &entity.Session{
				ID:        "S1",
				WalletID:  "W1",
				Token:     token,
				IsValid:   true,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	interceptor := NewGRPCAuthInterceptor(auth)

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	// create context with authorization metadata
	md := metadata.Pairs("authorization", "Bearer my-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/wallet.v1.WalletService/Transfer"}
	_, err := interceptor.UnaryInterceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("UnaryInterceptor() error = %v", err)
	}

	// verify context has wallet ID
	walletID := GetWalletID(capturedCtx)
	if walletID != "W1" {
		t.Errorf("walletID in context = %q, want %q", walletID, "W1")
	}

	// verify context has session token
	sessionToken := GetSessionToken(capturedCtx)
	if sessionToken != "my-token" {
		t.Errorf("sessionToken in context = %q, want %q", sessionToken, "my-token")
	}
}

func TestGRPCAuthInterceptor_AuthenticatedMethod_NoToken(t *testing.T) {
	auth := &mockAuthService{}
	interceptor := NewGRPCAuthInterceptor(auth)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Error("handler should not be called when no token provided")
		return nil, nil
	}

	// no metadata at all
	info := &grpc.UnaryServerInfo{FullMethod: "/wallet.v1.WalletService/Transfer"}
	_, err := interceptor.UnaryInterceptor(context.Background(), nil, info, handler)
	if err == nil {
		t.Fatal("UnaryInterceptor() expected error for missing token")
	}
}

func TestGRPCAuthInterceptor_AuthenticatedMethod_InvalidToken(t *testing.T) {
	auth := &mockAuthService{
		validateSessionFn: func(ctx context.Context, token string) (*entity.Session, error) {
			return nil, errors.ErrInvalidToken
		},
	}
	interceptor := NewGRPCAuthInterceptor(auth)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Error("handler should not be called for invalid token")
		return nil, nil
	}

	md := metadata.Pairs("authorization", "Bearer bad-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/wallet.v1.WalletService/Transfer"}
	_, err := interceptor.UnaryInterceptor(ctx, nil, info, handler)
	if err == nil {
		t.Fatal("UnaryInterceptor() expected error for invalid token")
	}
}

// --- Tests for extractTokenFromMetadata ---

func TestExtractToken_BearerToken(t *testing.T) {
	md := metadata.Pairs("authorization", "Bearer my-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractTokenFromMetadata(ctx)
	if err != nil {
		t.Fatalf("extractTokenFromMetadata() error = %v", err)
	}
	if token != "my-token" {
		t.Errorf("token = %q, want %q", token, "my-token")
	}
}

func TestExtractToken_BearerLowercase(t *testing.T) {
	md := metadata.Pairs("authorization", "bearer my-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractTokenFromMetadata(ctx)
	if err != nil {
		t.Fatalf("extractTokenFromMetadata() error = %v", err)
	}
	if token != "my-token" {
		t.Errorf("token = %q, want %q", token, "my-token")
	}
}

func TestExtractToken_RawToken(t *testing.T) {
	md := metadata.Pairs("authorization", "raw-token-value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractTokenFromMetadata(ctx)
	if err != nil {
		t.Fatalf("extractTokenFromMetadata() error = %v", err)
	}
	if token != "raw-token-value" {
		t.Errorf("token = %q, want %q", token, "raw-token-value")
	}
}

func TestExtractToken_XSessionTokenHeader(t *testing.T) {
	md := metadata.Pairs("x-session-token", "session-token-value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractTokenFromMetadata(ctx)
	if err != nil {
		t.Fatalf("extractTokenFromMetadata() error = %v", err)
	}
	if token != "session-token-value" {
		t.Errorf("token = %q, want %q", token, "session-token-value")
	}
}

func TestExtractToken_NoMetadata(t *testing.T) {
	_, err := extractTokenFromMetadata(context.Background())
	if err == nil {
		t.Fatal("extractTokenFromMetadata() expected error for no metadata")
	}
}

func TestExtractToken_EmptyMetadata(t *testing.T) {
	md := metadata.New(nil)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := extractTokenFromMetadata(ctx)
	if err == nil {
		t.Fatal("extractTokenFromMetadata() expected error for empty metadata")
	}
}

// --- Tests for context helpers ---

func TestGetWalletID_FromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), walletIDKey, "W1")
	if got := GetWalletID(ctx); got != "W1" {
		t.Errorf("GetWalletID() = %q, want %q", got, "W1")
	}
}

func TestGetWalletID_NotSet(t *testing.T) {
	if got := GetWalletID(context.Background()); got != "" {
		t.Errorf("GetWalletID() = %q, want empty", got)
	}
}

func TestGetSessionToken_FromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), sessionTokenKey, "token-123")
	if got := GetSessionToken(ctx); got != "token-123" {
		t.Errorf("GetSessionToken() = %q, want %q", got, "token-123")
	}
}

func TestGetSessionToken_NotSet(t *testing.T) {
	if got := GetSessionToken(context.Background()); got != "" {
		t.Errorf("GetSessionToken() = %q, want empty", got)
	}
}
