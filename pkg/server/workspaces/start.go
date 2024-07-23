// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/internal/util"
)

func (s *WorkspaceService) StartWorkspace(ctx context.Context, workspaceId string) error {
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

	err = s.startWorkspace(ctx, w, target, wsLogWriter)

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w, target)
	event := telemetry.ServerEventWorkspaceStarted
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceStartError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

func (s *WorkspaceService) StartProject(ctx context.Context, workspaceId, projectName string) error {
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

	return s.startProject(ctx, project, target, projectLogger)
}

func (s *WorkspaceService) startWorkspace(ctx context.Context, ws *workspace.Workspace, target *provider.ProviderTarget, wsLogWriter io.Writer) error {
	wsLogWriter.Write([]byte("Starting workspace\n"))

	ws.EnvVars = workspace.GetWorkspaceEnvVars(ws, workspace.WorkspaceEnvVarParams{
		ApiUrl:    s.serverApiUrl,
		ServerUrl: s.serverUrl,
		ClientId:  telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err := s.provisioner.StartWorkspace(ws, target)
	if err != nil {
		return err
	}

	for _, project := range ws.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(ws.Id, project.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		err = s.startProject(ctx, project, target, projectLogger)
		if err != nil {
			return err
		}
	}

	wsLogWriter.Write([]byte(fmt.Sprintf("Workspace %s started\n", ws.Name)))

	return nil
}

func (s *WorkspaceService) startProject(ctx context.Context, p *project.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Starting project %s\n", p.Name)))

	projectToStart := *p
	projectToStart.EnvVars = project.GetProjectEnvVars(p, project.ProjectEnvVarParams{
		ApiUrl:    s.serverApiUrl,
		ServerUrl: s.serverUrl,
		ClientId:  telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err := s.provisioner.StartProject(p, target)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s started\n", p.Name)))

	return nil
}
