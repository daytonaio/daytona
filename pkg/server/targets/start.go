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
	"github.com/daytonaio/daytona/pkg/target/project"
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

func (s *TargetService) StartProject(ctx context.Context, targetId, projectName string) error {
	w, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	project, err := w.GetProject(projectName)
	if err != nil {
		return ErrProjectNotFound
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &w.TargetConfig})
	if err != nil {
		return err
	}

	projectLogger := s.loggerFactory.CreateProjectLogger(w.Id, project.Name, logs.LogSourceServer)
	defer projectLogger.Close()

	return s.startProject(ctx, project, targetConfig, projectLogger)
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

	for _, project := range t.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(t.Id, project.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		err = s.startProject(ctx, project, targetConfig, projectLogger)
		if err != nil {
			return err
		}
	}

	targetLogger.Write([]byte(fmt.Sprintf("Target %s started\n", t.Name)))

	return nil
}

func (s *TargetService) startProject(ctx context.Context, p *project.Project, targetConfig *provider.TargetConfig, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Starting project %s\n", p.Name)))

	projectToStart := *p
	projectToStart.EnvVars = project.GetProjectEnvVars(p, project.ProjectEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	cr, err := s.containerRegistryService.FindByImageName(p.Image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	builderCr, err := s.containerRegistryService.FindByImageName(s.builderImage)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	var gc *gitprovider.GitProviderConfig

	if p.GitProviderConfigId != nil {
		gc, err = s.gitProviderService.GetConfig(*p.GitProviderConfigId)
		if err != nil && !gitprovider.IsGitProviderNotFound(err) {
			return err
		}
	}

	err = s.provisioner.StartProject(provisioner.ProjectParams{
		Project:                       &projectToStart,
		TargetConfig:                  targetConfig,
		ContainerRegistry:             cr,
		GitProviderConfig:             gc,
		BuilderImage:                  s.builderImage,
		BuilderImageContainerRegistry: builderCr,
	})
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s started\n", p.Name)))

	return nil
}
