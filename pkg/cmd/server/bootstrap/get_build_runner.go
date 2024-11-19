// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/scheduler"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

func GetBuildRunner(c *server.Config, buildRunnerConfig *build.Config, telemetryService telemetry.TelemetryService) (*build.BuildRunner, error) {
	logsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}
	loggerFactory := logs.NewLoggerFactory(nil, &logsDir)

	dbPath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	dbConnection := db.GetSQLiteConnection(dbPath)

	gitProviderConfigStore, err := db.NewGitProviderConfigStore(dbConnection)
	if err != nil {
		return nil, err
	}

	gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
		ConfigStore: gitProviderConfigStore,
	})

	buildStore, err := db.NewBuildStore(dbConnection)
	if err != nil {
		return nil, err
	}

	buildImageNamespace := c.BuildImageNamespace
	if buildImageNamespace != "" {
		buildImageNamespace = fmt.Sprintf("/%s", buildImageNamespace)
	}
	buildImageNamespace = strings.TrimSuffix(buildImageNamespace, "/")

	containerRegistryStore, err := db.NewContainerRegistryStore(dbConnection)
	if err != nil {
		return nil, err
	}

	containerRegistryService := containerregistries.NewContainerRegistryService(containerregistries.ContainerRegistryServiceConfig{
		Store: containerRegistryStore,
	})

	var builderRegistry *models.ContainerRegistry

	if c.BuilderRegistryServer != "local" {
		builderRegistry, err = containerRegistryService.Find(c.BuilderRegistryServer)
		if err != nil {
			builderRegistry = &models.ContainerRegistry{
				Server: c.BuilderRegistryServer,
			}
		}
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	builderFactory := build.NewBuilderFactory(build.BuilderFactoryConfig{
		Image:                 c.BuilderImage,
		ContainerRegistry:     builderRegistry,
		BuildStore:            buildStore,
		BuildImageNamespace:   buildImageNamespace,
		LoggerFactory:         loggerFactory,
		DefaultWorkspaceImage: c.DefaultWorkspaceImage,
		DefaultWorkspaceUser:  c.DefaultWorkspaceUser,
	})

	return build.NewBuildRunner(build.BuildRunnerInstanceConfig{
		Interval:          buildRunnerConfig.Interval,
		Scheduler:         scheduler.NewCronScheduler(),
		BuildRunnerId:     buildRunnerConfig.Id,
		ContainerRegistry: builderRegistry,
		GitProviderStore:  gitProviderService,
		BuildStore:        buildStore,
		BuilderFactory:    builderFactory,
		LoggerFactory:     loggerFactory,
		BasePath:          filepath.Join(configDir, "builds"),
		TelemetryService:  telemetryService,
	}), nil
}
