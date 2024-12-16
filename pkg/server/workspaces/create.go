// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"regexp"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req services.CreateWorkspaceDTO) (*services.WorkspaceDTO, error) {
	var err error
	ctx, err = s.workspaceStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.workspaceStore)

	_, err = s.workspaceStore.Find(ctx, req.Name)
	if err == nil {
		return s.handleCreateError(ctx, nil, services.ErrWorkspaceAlreadyExists)
	}

	target, err := s.findTarget(ctx, req.TargetId)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	w := req.ToWorkspace()
	w.Target = *target

	if !isValidWorkspaceName(w.Name) {
		return s.handleCreateError(ctx, w, services.ErrInvalidWorkspaceName)
	}

	w.Repository.Url = util.CleanUpRepositoryUrl(w.Repository.Url)

	if w.GitProviderConfigId == nil || *w.GitProviderConfigId == "" {
		configs, err := s.listGitProviderConfigs(ctx, w.Repository.Url)
		if err != nil {
			return s.handleCreateError(ctx, w, err)
		}

		if len(configs) > 1 {
			return s.handleCreateError(ctx, w, errors.New("multiple git provider configs found for the repository url"))
		}

		if len(configs) == 1 {
			w.GitProviderConfigId = &configs[0].Id
		}
	}

	if w.GitProviderConfigId != nil && *w.GitProviderConfigId == "" {
		w.GitProviderConfigId = nil
	}

	if w.Repository.Sha == "" {
		sha, err := s.getLastCommitSha(ctx, w.Repository)
		if err != nil {
			return s.handleCreateError(ctx, w, err)
		}
		w.Repository.Sha = sha
	}

	if w.BuildConfig != nil {
		cachedBuild, err := s.findCachedBuild(ctx, w)
		if err == nil {
			w.BuildConfig.CachedBuild = cachedBuild
		}
	}

	if w.Image == "" {
		w.Image = s.defaultWorkspaceImage
	}

	if w.User == "" {
		w.User = s.defaultWorkspaceUser
	}

	apiKey, err := s.generateApiKey(ctx, w.Id)
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	w.ApiKey = apiKey

	daytonaWorkspaceEnvVars := GetWorkspaceEnvVars(w, WorkspaceEnvVarParams{
		ApiUrl:           s.serverApiUrl,
		ServerUrl:        s.serverUrl,
		ServerVersion:    s.serverVersion,
		ClientId:         telemetry.ClientId(ctx),
		TelemetryEnabled: telemetry.TelemetryEnabled(ctx),
	})
	w.EnvVars = util.MergeEnvVars(daytonaWorkspaceEnvVars, w.EnvVars)

	err = s.workspaceStore.Save(ctx, w)
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	err = s.workspaceMetadataStore.Save(ctx, &models.WorkspaceMetadata{
		WorkspaceId: w.Id,
		Uptime:      0,
		GitStatus:   &models.GitStatus{},
	})
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionCreate)
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleCreateError(ctx, w, err)
}

func (s *WorkspaceService) handleCreateError(ctx context.Context, w *models.Workspace, err error) (*services.WorkspaceDTO, error) {
	if err != nil {
		err = s.workspaceStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		if w == nil {
			return nil, err
		}
		return &services.WorkspaceDTO{
			Workspace: *w,
			State:     w.GetState(),
		}, err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w)
	event := telemetry.ServerEventWorkspaceCreated
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceCreateError
	}
	telemetryError := s.trackTelemetryEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	if w == nil {
		return nil, err
	}

	return &services.WorkspaceDTO{
		Workspace: *w,
		State:     w.GetState(),
	}, err
}

func isValidWorkspaceName(name string) bool {
	// The repository name can only contain ASCII letters, digits, and the characters ., -, and _.
	var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Check if the name matches the basic regex
	if !validName.MatchString(name) {
		return false
	}

	// Names starting with a period must have atleast one char appended to it.
	if name == "." || name == "" {
		return false
	}

	return true
}
