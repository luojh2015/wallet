package middleware

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Module("ginmiddleware",
		fx.Provide(NewGinMiddleware),
	),
	fx.Module("grpcmiddleware",
		fx.Provide(NewGRPCAuthInterceptor),
	),
)
