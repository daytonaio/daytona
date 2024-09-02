// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"path/filepath"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	hstypes "github.com/juanfont/headscale/hscontrol/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
)

func (s *HeadscaleServer) getHeadscaleConfig() (*hstypes.Config, error) {
	cfg := &hstypes.Config{
		ServerURL:                      fmt.Sprintf("https://%s.%s", s.serverId, s.frpsDomain),
		Addr:                           fmt.Sprintf("0.0.0.0:%d", s.headscalePort),
		EphemeralNodeInactivityTimeout: 5 * time.Minute,
		BaseDomain:                     "daytona.local",
		DERP: hstypes.DERPConfig{
			ServerEnabled:                      true,
			AutomaticallyAddEmbeddedDerpRegion: true,
			ServerRegionID:                     999,
			ServerRegionCode:                   "local",
			ServerRegionName:                   "Daytona embedded DERP",
			Paths:                              []string{},
			ServerPrivateKeyPath:               filepath.Join(s.configDir, "derp_server_private.key"),
			UpdateFrequency:                    24 * time.Hour,
			AutoUpdate:                         true,
			STUNAddr:                           "0.0.0.0:3478",
		},
		Log: hstypes.LogConfig{
			Format: "text",
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

		Database: hstypes.DatabaseConfig{
			Sqlite: hstypes.SqliteConfig{
				Path: filepath.Join(s.configDir, "headscale.db"),
			},
			Type: "sqlite3",
		},
		UnixSocket:           filepath.Join(s.configDir, "headscale.sock"),
		UnixSocketPermission: fs.FileMode.Perm(0700),
		NoisePrivateKeyPath:  filepath.Join(s.configDir, "noise_private.key"),
		CLI: hstypes.CLIConfig{
			Timeout: 10 * time.Second,
		},
		Tuning: hstypes.Tuning{
			BatchChangeDelay:               800 * time.Millisecond,
			NotifierSendTimeout:            800 * time.Millisecond,
			NodeMapSessionBufferedChanSize: 30,
		},
		IPAllocation: hstypes.IPAllocationStrategySequential,
		Policy: hstypes.PolicyConfig{
			Mode: hstypes.PolicyModeDB,
		},
	}

	v4Prefix := netip.MustParsePrefix("100.64.0.0/10")
	v6Prefix := netip.MustParsePrefix("fd7a:115c:a1e0::/48")
	cfg.PrefixV4 = &v4Prefix
	cfg.PrefixV6 = &v6Prefix

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")
	if logLevelSet {
		var err error
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
