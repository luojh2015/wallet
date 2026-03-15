package grpc

import "go.uber.org/fx"

var Module = fx.Module("grpchandler",
	fx.Provide(NewWalletServiceServer),
)
