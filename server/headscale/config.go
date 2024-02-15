package headscale

import (
	"fmt"
	"os"
	"path"
	"time"

	proto_types "github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/daytonaio/daytona/plugins/utils"
	"github.com/daytonaio/daytona/server/config"
	"github.com/juanfont/headscale/hscontrol/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getConfig(serverConfig *proto_types.ServerConfig) (*types.Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
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

	cfg.ServerURL = fmt.Sprintf("http://127.0.0.1:%d", serverConfig.HeadscalePort)
	cfg.Addr = fmt.Sprintf("127.0.0.1:%d", serverConfig.HeadscalePort)

	cfg.DBpath = path.Join(configDir, "headscale", "headscale.db")
	cfg.UnixSocket = path.Join(configDir, "headscale", "headscale.sock")
	cfg.NoisePrivateKeyPath = path.Join(configDir, "headscale", "noise_private.key")
	cfg.DERP.ServerPrivateKeyPath = path.Join(configDir, "headscale", "derp_server_private.key")

	return cfg, nil
}

func init() {
	c, err := config.GetConfigDir()
	if err != nil {
		return
	}

	err = os.MkdirAll(path.Join(c, "headscale"), 0700)
	if err != nil {
		log.Error().Err(err).Msg("failed to create headscale directory")
	}
}
