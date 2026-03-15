package server

import (
	"fmt"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	v1 "github.com/luojh/wallet/api/grpc/v1"
	handler "github.com/luojh/wallet/internal/handler/grpc"
	"github.com/luojh/wallet/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// recoveryHandler 自定义 panic 恢复处理函数
func recoveryHandler(p interface{}) error {
	fmt.Printf("gRPC panic recovered: %v\n%s\n", p, string(debug.Stack()))
	return status.Errorf(codes.Internal, "internal server error")
}

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(
	walletServer *handler.WalletServiceServer,
	authInterceptor *middleware.GRPCAuthInterceptor,
) *grpc.Server {
	// 创建恢复拦截器
	recoveryInterceptor := recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(recoveryHandler))

	// 创建 gRPC 服务器，添加恢复器和认证拦截器
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor,
			authInterceptor.UnaryInterceptor,
		),
	)

	// 注册钱包服务
	v1.RegisterWalletServiceServer(server, walletServer)

	// 注册 gRPC 反射服务（用于 grpcurl 等工具）
	reflection.Register(server)

	return server
}
