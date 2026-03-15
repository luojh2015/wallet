package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/pkg/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var Module = fx.Module("server",
	fx.Provide(NewGRPCServer),
	fx.Provide(NewHTTPServer),
	fx.Invoke(startServers),
	fx.Invoke(registerHooks),
)

// startServers 启动服务器
func startServers(
	httpServer *http.Server,
	grpcServer *grpc.Server,
	cfg *config.Config,
) {
	// 启动 HTTP 服务器（包含 gRPC-Gateway REST API）
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// 启动 gRPC 服务器
	go func() {
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.App.GRPCPort))
		if err != nil {
			panic(err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func registerHooks(
	lc fx.Lifecycle,
	cfg *config.Config,
	httpServer *http.Server,
	grpcServer *grpc.Server,
	logger logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting wallet service...",
				zap.String("http_port", strconv.Itoa(cfg.App.HTTPPort)),
				zap.String("grpc_port", strconv.Itoa(cfg.App.GRPCPort)),
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping wallet service...")

			// 优雅关闭 HTTP 服务器
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to shutdown HTTP server", zap.Error(err))
			}

			// 优雅关闭 gRPC 服务器
			grpcServer.GracefulStop()

			return nil
		},
	})
}
