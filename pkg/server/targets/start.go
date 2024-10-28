// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/internal/util"
)

func (s *TargetService) StartTarget(ctx context.Context, targetId string) error {
	w, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &w.TargetConfig})
	if err != nil {
		return err
	}

	targetLogger := s.loggerFactory.CreateTargetLogger(w.Id, logs.LogSourceServer)
	defer targetLogger.Close()

	tgLogWriter := io.MultiWriter(&util.InfoLogWriter{}, targetLogger)

	err = s.startTarget(ctx, w, targetConfig, tgLogWriter)

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, w, targetConfig)
	event := telemetry.ServerEventTargetStarted
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetStartError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

func (s *TargetService) StartWorkspace(ctx context.Context, targetId, workspaceName string) error {
	w, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	workspace, err := w.GetWorkspace(workspaceName)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &w.TargetConfig})
	if err != nil {
		return err
	}

	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(w.Id, workspace.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	return s.startWorkspace(ctx, workspace, targetConfig, workspaceLogger)
}

func (s *TargetService) startTarget(ctx context.Context, t *target.Target, targetConfig *provider.TargetConfig, targetLogger io.Writer) error {
	targetLogger.Write([]byte("Starting target\n"))

	t.EnvVars = target.GetTargetEnvVars(t, target.TargetEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err := s.provisioner.StartTarget(t, targetConfig)
	if err != nil {
		return err
	}

	for _, workspace := range t.Workspaces {
		workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(t.Id, workspace.Name, logs.LogSourceServer)
		defer workspaceLogger.Close()

		err = s.startWorkspace(ctx, workspace, targetConfig, workspaceLogger)
		if err != nil {
			return err
		}
	}

	targetLogger.Write([]byte(fmt.Sprintf("Target %s started\n", t.Name)))

	return nil
}

func (s *TargetService) startWorkspace(ctx context.Context, w *workspace.Workspace, targetConfig *provider.TargetConfig, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Starting workspace %s\n", w.Name)))

	workspaceToStart := *w
	workspaceToStart.EnvVars = workspace.GetWorkspaceEnvVars(w, workspace.WorkspaceEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

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
		Workspace:                     &workspaceToStart,
		TargetConfig:                  targetConfig,
		ContainerRegistry:             cr,
		GitProviderConfig:             gc,
		BuilderImage:                  s.builderImage,
		BuilderImageContainerRegistry: builderCr,
	})
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Workspace %s started\n", w.Name)))

	return nil
}
