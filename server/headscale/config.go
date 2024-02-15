package headscale

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/daytonaio/daytona/plugins/utils"
	"github.com/juanfont/headscale/hscontrol/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getConfig() (*types.Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	err = types.LoadConfig(path.Join(pwd, "config.yaml"), true)
	if err != nil {
		return nil, fmt.Errorf("failed to load headscale configuration: %w", err)
	}

	cfg, err := types.GetHeadscaleConfig()
	if err != nil {
		return nil, err
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        &utils.DebugLogWriter{},
		TimeFormat: time.RFC3339,
	})

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")
	if logLevelSet {
		cfg.Log.Level, err = zerolog.ParseLevel(logLevelEnv)
		if err != nil {
			cfg.Log.Level = zerolog.ErrorLevel
		}
	} else {
		cfg.Log.Level = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(cfg.Log.Level)

	cfg.Log.Format = "text"

	return cfg, nil
}
