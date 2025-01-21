// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StartWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, ErrWorkspaceNotFound)
	}

	target, err := s.targetStore.Find(ws.TargetId)
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, err)
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, err)
	}

	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, ws.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	workspaceToStart := ws.Workspace
	workspaceToStart.EnvVars = workspace.GetWorkspaceEnvVars(&ws.Workspace, workspace.WorkspaceEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err = s.startWorkspace(&workspaceToStart, targetConfig, workspaceLogger)
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, err)
	}

	return s.handleStartError(ctx, &ws.Workspace, err)
}

func (s *WorkspaceService) startWorkspace(w *workspace.Workspace, targetConfig *provider.TargetConfig, logger io.Writer) error {
	logger.Write([]byte(fmt.Sprintf("Starting workspace %s\n", w.Name)))

	cr, err := s.containerRegistryService.FindByImageName(w.Image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	builderCr, err := s.containerRegistryService.FindByImageName(s.builderImage)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	var gc *gitprovider.GitProviderConfig

	if w.GitProviderConfigId != nil {
		gc, err = s.gitProviderService.GetConfig(*w.GitProviderConfigId)
		if err != nil && !gitprovider.IsGitProviderNotFound(err) {
			return err
		}
	}

	err = s.provisioner.StartWorkspace(provisioner.WorkspaceParams{
		Workspace:                     w,
		TargetConfig:                  targetConfig,
		ContainerRegistry:             cr,
		GitProviderConfig:             gc,
		BuilderImage:                  s.builderImage,
		BuilderImageContainerRegistry: builderCr,
	})
	if err != nil {
		return err
	}

	logger.Write([]byte(fmt.Sprintf("Workspace %s started\n", w.Name)))

	return nil
}

func (s *WorkspaceService) handleStartError(ctx context.Context, w *workspace.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w)
	event := telemetry.ServerEventWorkspaceStarted
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceStartError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
