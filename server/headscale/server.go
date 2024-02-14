package headscale

import (
	"fmt"
	"os"
	"path"

	"github.com/juanfont/headscale/hscontrol"
	"github.com/juanfont/headscale/hscontrol/types"

	log "github.com/sirupsen/logrus"
)

func Start() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	log.Info(pwd)

	err = types.LoadConfig(path.Join(pwd, "config.yaml"), true)
	if err != nil {
		return fmt.Errorf("failed to load headscale configuration: %w", err)
	}

	cfg, err := types.GetHeadscaleConfig()
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
