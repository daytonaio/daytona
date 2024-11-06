// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/buildconfig"

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

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req dto.CreateWorkspaceDTO) (*workspace.WorkspaceViewDTO, error) {
	_, err := s.workspaceStore.Find(req.Name)
	if err == nil {
		return s.handleCreateError(ctx, nil, ErrWorkspaceAlreadyExists)
	}

	target, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &req.TargetId})
	if err != nil {
		return s.handleCreateError(ctx, nil, err)
	}

	w := conversion.CreateDtoToWorkspace(req)

	if !isValidWorkspaceName(w.Name) {
		return s.handleCreateError(ctx, w, ErrInvalidWorkspaceName)
	}

	w.Repository.Url = util.CleanUpRepositoryUrl(w.Repository.Url)
	if w.GitProviderConfigId == nil || *w.GitProviderConfigId == "" {
		configs, err := s.gitProviderService.ListConfigsForUrl(w.Repository.Url)
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
		sha, err := s.gitProviderService.GetLastCommitSha(w.Repository)
		if err != nil {
			return s.handleCreateError(ctx, w, err)
		}
		w.Repository.Sha = sha
	}

	if w.BuildConfig != nil {
		cachedBuild, err := s.getCachedBuildForWorkspace(w)
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

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, fmt.Sprintf("ws-%s", w.Id))
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
	workspaceWithEnv.EnvVars = workspace.GetWorkspaceEnvVars(w, workspace.WorkspaceEnvVarParams{
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

	cr, err := s.containerRegistryService.FindByImageName(w.Image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return s.handleCreateError(ctx, w, err)
	}

	builderCr, err := s.containerRegistryService.FindByImageName(s.builderImage)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return s.handleCreateError(ctx, w, err)
	}

	var gc *gitprovider.GitProviderConfig

	if w.GitProviderConfigId != nil {
		gc, err = s.gitProviderService.GetConfig(*w.GitProviderConfigId)
		if err != nil && !gitprovider.IsGitProviderNotFound(err) {
			return s.handleCreateError(ctx, w, err)
		}
	}

	err = s.provisioner.CreateWorkspace(provisioner.WorkspaceParams{
		Workspace:                     w,
		Target:                        &target.Target,
		ContainerRegistry:             cr,
		GitProviderConfig:             gc,
		BuilderImage:                  s.builderImage,
		BuilderImageContainerRegistry: builderCr,
	})
	if err != nil {
		return s.handleCreateError(ctx, w, err)
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s created", w.Name))))

	err = s.startWorkspace(w, &target.Target, workspaceLogger)

	return s.handleCreateError(ctx, w, err)
}

func (s *WorkspaceService) handleCreateError(ctx context.Context, w *workspace.Workspace, err error) (*workspace.WorkspaceViewDTO, error) {
	if !telemetry.TelemetryEnabled(ctx) {
		if w == nil {
			return nil, err
		}
		return &workspace.WorkspaceViewDTO{Workspace: *w}, err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w)
	event := telemetry.ServerEventWorkspaceCreated
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceCreateError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	if w == nil {
		return nil, err
	}

	return &workspace.WorkspaceViewDTO{Workspace: *w}, err
}

func (s *WorkspaceService) getCachedBuildForWorkspace(w *workspace.Workspace) (*buildconfig.CachedBuild, error) {
	validStates := &[]build.BuildState{
		build.BuildState(build.BuildStatePublished),
	}

	build, err := s.buildService.Find(&build.Filter{
		States:        validStates,
		RepositoryUrl: &w.Repository.Url,
		Branch:        &w.Repository.Branch,
		EnvVars:       &w.EnvVars,
		BuildConfig:   w.BuildConfig,
		GetNewest:     util.Pointer(true),
	})
	if err != nil {
		return nil, err
	}

	if build.Image == nil || build.User == nil {
		return nil, errors.New("cached build is missing image or user")
	}

	return &buildconfig.CachedBuild{
		User:  *build.User,
		Image: *build.Image,
	}, nil
}
