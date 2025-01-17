// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/pkg/stringid"

	log "github.com/sirupsen/logrus"
)

func (s *BuildService) Create(ctx context.Context, b services.CreateBuildDTO) (string, error) {
	ctx, err := s.buildStore.BeginTransaction(ctx)
	if err != nil {
		return "", s.handleCreateError(ctx, nil, err)
	}

	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)

	workspaceTemplate, err := s.findWorkspaceTemplate(ctx, b.WorkspaceTemplateName)
	if err != nil {
		return "", s.handleCreateError(ctx, nil, err)
	}

	repo, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl, b.Branch)
	if err != nil {
		return "", s.handleCreateError(ctx, nil, err)
	}

	newBuild := models.Build{
		Id: id,
		ContainerConfig: models.ContainerConfig{
			Image: workspaceTemplate.Image,
			User:  workspaceTemplate.User,
		},
		BuildConfig: workspaceTemplate.BuildConfig,
		Repository:  repo,
		EnvVars:     b.EnvVars,
	}

	if b.PrebuildId != nil {
		newBuild.PrebuildId = b.PrebuildId
	}

	err = s.buildStore.Save(ctx, &newBuild)
	if err != nil {
		return "", s.handleCreateError(ctx, nil, err)
	}

	err = s.createJob(ctx, id, models.JobActionRun)
	if err != nil {
		return "", s.handleCreateError(ctx, &newBuild, err)
	}

	err = s.buildStore.CommitTransaction(ctx)
	return id, s.handleCreateError(ctx, &newBuild, err)
}

func (s *BuildService) handleCreateError(ctx context.Context, b *models.Build, err error) error {
	if err != nil {
		err = s.buildStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.BuildEventLifecycleCreated
	if err != nil {
		eventName = telemetry.BuildEventLifecycleCreationFailed
	}
	event := telemetry.NewBuildEvent(eventName, b, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
