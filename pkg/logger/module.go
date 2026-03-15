package logger

import (
	"github.com/luojh/wallet/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Module("zag_logger", fx.Provide(func(cfg *config.Config) (Logger, error) {
	return NewLogger(cfg.App.Env)
}))
