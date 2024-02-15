package headscale

import (
	"fmt"

	"github.com/juanfont/headscale/hscontrol"
)

func Start() error {
	cfg, err := getConfig()
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
