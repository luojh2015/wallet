package middleware

import (
	"context"
	"strings"

	"github.com/luojh/wallet/internal/service"
	"github.com/luojh/wallet/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// AuthorizationHeader Authorization 头
	AuthorizationHeader = "authorization"
	// WalletIDKey 上下文中钱包ID的键
	WalletIDKey = "wallet_id"
	// SessionTokenKey 上下文中会话令牌的键
	SessionTokenKey = "session_token"
)

// contextKey 上下文键类型
type contextKey string

const (
	walletIDKey     contextKey = "wallet_id"
	sessionTokenKey contextKey = "session_token"
)

// GRPCAuthInterceptor gRPC 认证拦截器
type GRPCAuthInterceptor struct {
	authService service.IAuthService
	// 需要认证的方法
	authMethods map[string]bool
}

// NewGRPCAuthInterceptor 创建 gRPC 认证拦截器
func NewGRPCAuthInterceptor(authService service.IAuthService) *GRPCAuthInterceptor {
	return &GRPCAuthInterceptor{
		authService: authService,
		authMethods: map[string]bool{
			"/wallet.v1.WalletService/Transfer": true,
		},
	}
}

// UnaryInterceptor 一元 RPC 拦截器
func (i *GRPCAuthInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 检查是否需要认证
	if _, ok := i.authMethods[info.FullMethod]; !ok {
		return handler(ctx, req)
	}

	// 执行认证
	newCtx, err := i.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return handler(newCtx, req)
}

// authenticate 执行认证
func (i *GRPCAuthInterceptor) authenticate(ctx context.Context) (context.Context, error) {
	// 从 metadata 获取 token
	token, err := extractTokenFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// 验证会话
	session, err := i.authService.ValidateSession(ctx, token)
	if err != nil {
		appErr := errors.GetAppError(err)
		return nil, status.Errorf(codes.Unauthenticated, "%s", appErr.Message)
	}

	// 将认证信息存入 context
	ctx = context.WithValue(ctx, walletIDKey, session.WalletID)
	ctx = context.WithValue(ctx, sessionTokenKey, token)

	return ctx, nil
}

// extractTokenFromMetadata 从 gRPC metadata 提取 token
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "未提供认证令牌")
	}

	// 从 authorization header 获取
	values := md.Get(AuthorizationHeader)
	if len(values) > 0 {
		auth := values[0]
		// Bearer token
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer "), nil
		}
		if strings.HasPrefix(auth, "bearer ") {
			return strings.TrimPrefix(auth, "bearer "), nil
		}
		return auth, nil
	}

	// 从 x-session-token header 获取（gRPC-Gateway 传递）
	values = md.Get("x-session-token")
	if len(values) > 0 {
		return values[0], nil
	}

	return "", status.Error(codes.Unauthenticated, "未提供认证令牌")
}

// GetWalletID 从 context 获取钱包 ID
func GetWalletID(ctx context.Context) string {
	if walletID, ok := ctx.Value(walletIDKey).(string); ok {
		return walletID
	}
	return ""
}

// GetSessionToken 从 context 获取会话令牌
func GetSessionToken(ctx context.Context) string {
	if token, ok := ctx.Value(sessionTokenKey).(string); ok {
		return token
	}
	return ""
}
