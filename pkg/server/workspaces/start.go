// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"

	"github.com/daytonaio/daytona/internal/util"
)

func (s *WorkspaceService) StartWorkspace(workspaceId string) error {
	w, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	target, err := s.targetStore.Find(w.Target)
	if err != nil {
		return err
	}

	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(w.Id, logs.LogSourceServer)
	defer workspaceLogger.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, workspaceLogger)

	return s.startWorkspace(w, target, wsLogWriter)
}

func (s *WorkspaceService) StartProject(workspaceId, projectName string) error {
	w, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	project, err := w.GetProject(projectName)
	if err != nil {
		return ErrProjectNotFound
	}

	target, err := s.targetStore.Find(project.Target)
	if err != nil {
		return err
	}

	projectLogger := s.loggerFactory.CreateProjectLogger(w.Id, project.Name, logs.LogSourceServer)
	defer projectLogger.Close()

	return s.startProject(project, target, projectLogger)
}

func (s *WorkspaceService) startWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget, wsLogWriter io.Writer) error {
	wsLogWriter.Write([]byte("Starting workspace\n"))

	err := s.provisioner.StartWorkspace(workspace, target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(workspace.Id, project.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		err = s.startProject(project, target, projectLogger)
		if err != nil {
			return err
		}
	}

	wsLogWriter.Write([]byte(fmt.Sprintf("Workspace %s started\n", workspace.Name)))

	return nil
}

func (s *WorkspaceService) startProject(project *workspace.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Starting project %s\n", project.Name)))

	projectToStart := *project
	projectToStart.EnvVars = workspace.GetProjectEnvVars(project, s.serverApiUrl, s.serverUrl)

	err := s.provisioner.StartProject(project, target)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s started\n", project.Name)))

	return nil
}
