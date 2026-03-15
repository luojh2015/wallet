package wallet

import "go.uber.org/fx"

var Module = fx.Provide(NewWallet)
