package headscale

import (
	"fmt"
	"os"
	"path"
	"time"

	server_types "github.com/daytonaio/daytona/common/types"
	"github.com/daytonaio/daytona/plugins/utils"
	"github.com/daytonaio/daytona/server/config"
	"github.com/juanfont/headscale/hscontrol/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func getConfig(serverConfig *server_types.ServerConfig) (*types.Config, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	err = types.LoadConfig(path.Join(configDir, "headscale", "config.yaml"), true)
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
		return
	}

	if _, err := os.Stat(path.Join(c, "headscale", "config.yaml")); os.IsNotExist(err) {
		yamlString, err := yaml.Marshal(defaultConfig)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal default headscale configuration")
			return
		}

		err = os.WriteFile(path.Join(c, "headscale", "config.yaml"), yamlString, 0600)
		if err != nil {
			log.Error().Err(err).Msg("failed to write default headscale configuration")
			return
		}
	}
}

var defaultConfig map[string]interface{} = map[string]interface{}{
	"acl_policy_path": " ",
	"acme_email":      " ",
	"acme_url":        "https://acme-v02.api.letsencrypt.org/directory",
	"db_type":         "sqlite3",
	"derp": map[string]interface{}{
		"auto_update_enabled": true,
		"paths":               []interface{}{},
		"server": map[string]interface{}{
			"automatically_add_embedded_derp_region": true,
			"enabled":                                true,
			"ipv4":                                   "1.2.3.4",
			"ipv6":                                   "2001:db8::1",
			"region_code":                            "headscale",
			"region_id":                              999,
			"region_name":                            "Headscale Embedded DERP",
			"stun_listen_addr":                       "0.0.0.0:3478",
		},
		"update_frequency": "24h",
		"urls":             []interface{}{"https://controlplane.tailscale.com/derpmap/default"},
	},
	"disable_check_updates": false,
	"dns_config": map[string]interface{}{
		"base_domain":        "example.com",
		"domains":            []interface{}{},
		"magic_dns":          true,
		"nameservers":        []interface{}{"1.1.1.1"},
		"override_local_dns": true,
	},
	"ephemeral_node_inactivity_timeout": "30m",
	"grpc_allow_insecure":               false,
	"grpc_listen_addr":                  "127.0.0.1:50443",
	"ip_prefixes": []interface{}{
		"fd7a:115c:a1e0::/48",
		"100.64.0.0/10",
	},
	"log": map[string]interface{}{
		"format": "text",
		"level":  "info",
	},
	"logtail": map[string]interface{}{
		"enabled": false,
	},
	"metrics_listen_addr":        "127.0.0.1:9090",
	"node_update_check_interval": "10s",
	"noise": map[string]interface{}{
		"private_key_path": "/",
	},
	"randomize_client_port":          false,
	"server_url":                     "http://127.0.0.1:8000",
	"tls_cert_path":                  "",
	"tls_key_path":                   "",
	"tls_letsencrypt_cache_dir":      "/var/lib/headscale/cache",
	"tls_letsencrypt_challenge_type": "HTTP-01",
	"tls_letsencrypt_hostname":       "",
	"tls_letsencrypt_listen":         ":http",
}
