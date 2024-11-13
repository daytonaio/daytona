// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"

	log "github.com/sirupsen/logrus"
)

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

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req services.CreateWorkspaceDTO) (*models.Workspace, error) {
	_, err := s.workspaceStore.Find(req.Name)
	if err == nil {
		return s.handleCreateError(ctx, nil, ErrWorkspaceAlreadyExists)
	}

	target, err := s.findTarget(ctx, req.TargetId)
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	w := req.ToWorkspace()
	w.Target = *target

	if !isValidWorkspaceName(w.Name) {
		return s.handleCreateError(ctx, w, ErrInvalidWorkspaceName)
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

	apiKey, err := s.generateApiKey(ctx, fmt.Sprintf("ws-%s", w.Id))
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	w.ApiKey = apiKey

	err = s.workspaceStore.Save(w)
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(w.Id, w.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	workspaceWithEnv := *w
	workspaceWithEnv.EnvVars = GetWorkspaceEnvVars(w, WorkspaceEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	for k, v := range w.EnvVars {
		workspaceWithEnv.EnvVars[k] = v
	}

	w = &workspaceWithEnv

	workspaceLogger.Write([]byte(fmt.Sprintf("Creating workspace %s\n", w.Name)))

	cr, err := s.findContainerRegistry(ctx, w.Image)
	if err != nil && !stores.IsContainerRegistryNotFound(err) {
		return s.handleCreateError(ctx, w, err)
	}

	var gc *models.GitProviderConfig

	if w.GitProviderConfigId != nil {
		gc, err = s.findGitProviderConfig(ctx, *w.GitProviderConfigId)
		if err != nil && !stores.IsGitProviderNotFound(err) {
			return s.handleCreateError(ctx, w, err)
		}
	}

	err = s.provisioner.CreateWorkspace(w, cr, gc)
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s created", w.Name))))

	err = s.startWorkspace(w, workspaceLogger)

	return s.handleCreateError(ctx, w, err)
}

func (s *WorkspaceService) handleCreateError(ctx context.Context, w *models.Workspace, err error) (*models.Workspace, error) {
	if !telemetry.TelemetryEnabled(ctx) {
		if w == nil {
			return nil, err
		}
		return w, err
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

	return w, err
}
