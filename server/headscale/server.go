package headscale

import (
	"fmt"

	"github.com/daytonaio/daytona/common/types"
	"github.com/juanfont/headscale/hscontrol"
)

func Start(serverConfig *types.ServerConfig) error {
	cfg, err := getConfig(serverConfig)
	if err != nil {
		return fmt.Errorf(
			"failed to load configuration while creating headscale instance: %w",
			err,
		)
	}

	app, err := hscontrol.NewHeadscale(cfg)
	if err != nil {
		return err
	}

	return app.Serve()
}
