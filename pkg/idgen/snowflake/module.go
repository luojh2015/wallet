package snowflake

import (
	"github.com/luojh/wallet/internal/config"
	"github.com/luojh/wallet/pkg/idgen"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Module("snowflake",
		fx.Provide(func(cfg *config.Config) (idgen.IDGenerator, error) {
			return NewSnowflakeGenerator(int64(cfg.App.MachineID))
		}),
	),
)
