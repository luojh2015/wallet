package session

import "go.uber.org/fx"

var Module = fx.Provide(NewSessionDomain)
