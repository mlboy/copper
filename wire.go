// +build wireinject

package copper

import (
	"github.com/gocopper/copper/cconfig"
	"github.com/gocopper/copper/clogger"
	"github.com/google/wire"
)

// InitApp creates a new Copper app along with its dependencies.
func InitApp() (*App, error) {
	panic(
		wire.Build(
			NewApp,
			NewFlags,
			NewLifecycle,
			cconfig.New,
			clogger.NewWithConfig,

			wire.FieldsOf(new(*Flags), "Env", "ConfigDir", "ProjectDir"),
		),
	)
}

// WireModule can be used as part of google/wire setup to include the app's
// lifecycle, config, and logger.
var WireModule = wire.NewSet(
	wire.FieldsOf(new(*App), "Lifecycle", "Config", "Logger"),
)
