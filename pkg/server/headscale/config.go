// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"path"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/types"
	hstypes "github.com/juanfont/headscale/hscontrol/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
)

func getConfig(serverConfig *types.ServerConfig) (*hstypes.Config, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	cfg := &hstypes.Config{
		DBtype:                         "sqlite3",
		ServerURL:                      fmt.Sprintf("https://%s.%s", serverConfig.Id, serverConfig.Frps.Domain),
		Addr:                           fmt.Sprintf("127.0.0.1:%d", serverConfig.HeadscalePort),
		EphemeralNodeInactivityTimeout: 5 * time.Minute,
		NodeUpdateCheckInterval:        10 * time.Second,
		BaseDomain:                     "daytona.local",
		DERP: hstypes.DERPConfig{
			ServerEnabled:                      true,
			AutomaticallyAddEmbeddedDerpRegion: true,
			ServerRegionID:                     999,
			ServerRegionCode:                   "local",
			ServerRegionName:                   "Daytona embedded DERP",
			Paths:                              []string{},
			ServerPrivateKeyPath:               path.Join(configDir, "headscale", "derp_server_private.key"),
			UpdateFrequency:                    24 * time.Hour,
			AutoUpdate:                         true,
			STUNAddr:                           "0.0.0.0:3478",
		},
		Log: hstypes.LogConfig{
			Format: "text",
		},
		IPPrefixes: []netip.Prefix{
			netip.MustParsePrefix("fd7a:115c:a1e0::/48"),
			netip.MustParsePrefix("100.64.0.0/10"),
		},
		DNSConfig: &tailcfg.DNSConfig{
			Proxied: true,
			Nameservers: []netip.Addr{
				netip.MustParseAddr("127.0.0.11"),
				netip.MustParseAddr("1.1.1.1"),
			},
			Resolvers: []*dnstype.Resolver{
				{
					Addr: "127.0.0.11",
				},
				{
					Addr: "1.1.1.1",
				},
			},
		},
		DBpath:               path.Join(configDir, "headscale", "headscale.db"),
		UnixSocket:           path.Join(configDir, "headscale", "headscale.sock"),
		UnixSocketPermission: fs.FileMode.Perm(0700),
		NoisePrivateKeyPath:  path.Join(configDir, "headscale", "noise_private.key"),
		CLI: hstypes.CLIConfig{
			Timeout: 10 * time.Second,
		},
	}

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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        &util.DebugLogWriter{},
		TimeFormat: time.RFC3339,
	})

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
}
