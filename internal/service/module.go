package service

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Module("auth_service",
		fx.Provide(NewAuthService),
	),
	fx.Module("wallet_service",
		fx.Provide(NewWalletService),
	),
)
