// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/services"
)

func InitProviderManager(c *server.Config, configDir string) error {
	targetLogsDir, err := server.GetTargetLogsDir(configDir)
	if err != nil {
		return err
	}

	headscaleServer := headscale.NewHeadscaleServer(&headscale.HeadscaleServerConfig{
		ServerId:      c.Id,
		FrpsDomain:    c.Frps.Domain,
		FrpsProtocol:  c.Frps.Protocol,
		HeadscalePort: c.HeadscalePort,
		ConfigDir:     filepath.Join(configDir, "headscale"),
		Frps:          c.Frps,
	})
	err = headscaleServer.Init()
	if err != nil {
		return err
	}

	headscaleUrl := util.GetFrpcHeadscaleUrl(c.Frps.Protocol, c.Id, c.Frps.Domain)

	version := internal.Version

	dbPath, err := getDbPath()
	if err != nil {
		return err
	}

	dbConnection := db.GetSQLiteConnection(dbPath)

	store := db.NewStore(dbConnection)

	targetConfigStore, err := db.NewTargetConfigStore(store)
	if err != nil {
		return err
	}

	targetConfigService := targetconfigs.NewTargetConfigService(targetconfigs.TargetConfigServiceConfig{
		TargetConfigStore: targetConfigStore,
	})

	_ = manager.GetProviderManager(&manager.ProviderManagerConfig{
		LogsDir:            targetLogsDir,
		ApiUrl:             util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		DaytonaDownloadUrl: getDaytonaScriptUrl(c),
		ServerUrl:          headscaleUrl,
		ServerVersion:      version,
		RegistryUrl:        c.RegistryUrl,
		BaseDir:            c.ProvidersDir,
		CreateProviderNetworkKey: func(ctx context.Context, providerName string) (string, error) {
			return headscaleServer.CreateAuthKey(headscale.HEADSCALE_USERNAME)
		},
		ServerPort: c.HeadscalePort,
		ApiPort:    c.ApiPort,
		GetTargetConfigMap: func(ctx context.Context) (map[string]*models.TargetConfig, error) {
			return targetConfigService.Map(ctx)
		},
		CreateTargetConfig: func(ctx context.Context, name, options string, providerInfo models.ProviderInfo) error {
			_, err := targetConfigService.Add(ctx, services.AddTargetConfigDTO{
				Name:         name,
				Options:      options,
				ProviderInfo: providerInfo,
			})
			return err
		},
	})

	return nil
}

func getDaytonaScriptUrl(config *server.Config) string {
	url, _ := url.JoinPath(util.GetFrpcApiUrl(config.Frps.Protocol, config.Id, config.Frps.Domain), "binary", "script")
	return url
}
