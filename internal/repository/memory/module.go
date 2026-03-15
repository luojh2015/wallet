package memory

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Module("mem_session_repo",
		fx.Provide(NewSession),
	),
	fx.Module("mem_wallet_repo",
		fx.Provide(NewWallet),
	),
)
