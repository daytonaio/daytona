// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"

	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/db"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/types"
)

func StartWorkspace(workspaceId string) error {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	logsDir, err := config.GetWorkspaceLogsDir()
	if err != nil {
		return err
	}

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	target, err := targets.GetTarget(w.Target)
	if err != nil {
		return err
	}

	workspaceLogger := logger.GetWorkspaceLogger(logsDir, w.Id)
	defer workspaceLogger.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, workspaceLogger)

	return startWorkspace(w, target, c, logsDir, wsLogWriter)
}

func StartProject(workspaceId, projectId string) error {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	project, err := getProject(w, projectId)
	if err != nil {
		return errors.New("project not found")
	}

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	target, err := targets.GetTarget(w.Target)
	if err != nil {
		return err
	}

	logsDir, err := config.GetWorkspaceLogsDir()
	if err != nil {
		return err
	}

	workspaceLogger := logger.GetWorkspaceLogger(logsDir, w.Id)
	defer workspaceLogger.Close()

	projectLogger := logger.GetProjectLogger(logsDir, w.Id, project.Name)
	defer projectLogger.Close()

	projectLogWriter := io.MultiWriter(workspaceLogger, projectLogger)

	return startProject(project, target, c, projectLogWriter)
}

func startWorkspace(workspace *types.Workspace, target *provider.ProviderTarget, config *types.ServerConfig, logsDir string, wsLogWriter io.Writer) error {
	wsLogWriter.Write([]byte("Starting workspace\n"))

	err := provisioner.StartWorkspace(workspace, target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		projectLogger := logger.GetProjectLogger(logsDir, workspace.Id, project.Name)
		defer projectLogger.Close()

		projectLogWriter := io.MultiWriter(wsLogWriter, projectLogger)

		err = startProject(project, target, config, projectLogWriter)
		if err != nil {
			return err
		}
	}

	wsLogWriter.Write([]byte(fmt.Sprintf("Workspace %s started\n", workspace.Name)))

	return nil
}

func startProject(project *types.Project, target *provider.ProviderTarget, c *types.ServerConfig, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Starting project %s\n", project.Name)))

	projectToStart := *project
	projectToStart.EnvVars = getProjectEnvVars(project, c)

	err := provisioner.StartProject(project, target)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s started\n", project.Name)))

	return nil
}
