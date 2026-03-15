package main

import (
	"github.com/luojh/wallet/internal/config"
	dsession "github.com/luojh/wallet/internal/domain/session"
	dwallet "github.com/luojh/wallet/internal/domain/wallet"
	"github.com/luojh/wallet/internal/handler/grpc"
	"github.com/luojh/wallet/internal/middleware"
	"github.com/luojh/wallet/internal/repository/memory"
	"github.com/luojh/wallet/internal/server"
	"github.com/luojh/wallet/internal/service"
	"github.com/luojh/wallet/pkg/idgen"
	"github.com/luojh/wallet/pkg/idgen/snowflake"
	"github.com/luojh/wallet/pkg/logger"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(func() (*config.Config, error) { return config.Load("") }),
		logger.Module,
		snowflake.Module,
		fx.Provide(idgen.NewIDFactory),
		memory.Module,
		dsession.Module,
		dwallet.Module,
		middleware.Module,
		grpc.Module,
		service.Module,
		server.Module,
	).Run()
}
