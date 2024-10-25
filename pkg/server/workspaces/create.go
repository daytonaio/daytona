// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"

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

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req dto.CreateWorkspaceDTO) (*workspace.Workspace, error) {
	_, err := s.workspaceStore.Find(req.Name)
	if err == nil {
		return nil, ErrWorkspaceAlreadyExists
	}

	// Repo name is taken as the name for workspace by default
	if !isValidWorkspaceName(req.Name) {
		return nil, ErrInvalidWorkspaceName
	}

	w := &workspace.Workspace{
		Id:           req.Id,
		Name:         req.Name,
		TargetConfig: req.TargetConfig,
	}

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, w.Id)
	if err != nil {
		return nil, err
	}
	w.ApiKey = apiKey

	w.Projects = []*project.Project{}

	for _, projectDto := range req.Projects {
		p := conversion.CreateDtoToProject(projectDto)

		isValidProjectName := regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`).MatchString
		if !isValidProjectName(p.Name) {
			return nil, ErrInvalidProjectName
		}

		p.Repository.Url = util.CleanUpRepositoryUrl(p.Repository.Url)
		if p.GitProviderConfigId == nil || *p.GitProviderConfigId == "" {
			configs, err := s.gitProviderService.ListConfigsForUrl(p.Repository.Url)
			if err != nil {
				return nil, err
			}

			if len(configs) > 1 {
				return nil, errors.New("multiple git provider configs found for the repository url")
			}

			if len(configs) == 1 {
				p.GitProviderConfigId = &configs[0].Id
			}
		}

		if p.Repository.Sha == "" {
			sha, err := s.gitProviderService.GetLastCommitSha(p.Repository)
			if err != nil {
				return nil, err
			}
			p.Repository.Sha = sha
		}

		if p.BuildConfig != nil {
			cachedBuild, err := s.getCachedBuildForProject(p)
			if err == nil {
				p.BuildConfig.CachedBuild = cachedBuild
			}
		}

		if p.Image == "" {
			p.Image = s.defaultProjectImage
		}

		if p.User == "" {
			p.User = s.defaultProjectUser
		}

		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", w.Id, p.Name))
		if err != nil {
			return nil, err
		}

		p.WorkspaceId = w.Id
		p.ApiKey = apiKey
		p.TargetConfig = w.TargetConfig
		w.Projects = append(w.Projects, p)
	}

	err = s.workspaceStore.Save(w)
	if err != nil {
		return nil, err
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &w.TargetConfig})
	if err != nil {
		return w, err
	}

	w, err = s.createWorkspace(ctx, w, targetConfig)

	if !telemetry.TelemetryEnabled(ctx) {
		return w, err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w, targetConfig)
	event := telemetry.ServerEventWorkspaceCreated
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceCreateError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return w, err
}

func (s *WorkspaceService) createProject(p *project.Project, targetConfig *provider.TargetConfig, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", p.Name)))

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

	err = s.provisioner.CreateProject(provisioner.ProjectParams{
		Project:                       p,
		TargetConfig:                  targetConfig,
		ContainerRegistry:             cr,
		GitProviderConfig:             gc,
		BuilderImage:                  s.builderImage,
		BuilderImageContainerRegistry: builderCr,
	})
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s created\n", p.Name)))

	return nil
}

func (s *WorkspaceService) createWorkspace(ctx context.Context, ws *workspace.Workspace, targetConfig *provider.TargetConfig) (*workspace.Workspace, error) {
	wsLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, logs.LogSourceServer)
	defer wsLogger.Close()

	wsLogger.Write([]byte(fmt.Sprintf("Creating workspace %s (%s)\n", ws.Name, ws.Id)))

	ws.EnvVars = workspace.GetWorkspaceEnvVars(ws, workspace.WorkspaceEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err := s.provisioner.CreateWorkspace(ws, targetConfig)
	if err != nil {
		return nil, err
	}

	for i, p := range ws.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(ws.Id, p.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		projectWithEnv := *p
		projectWithEnv.EnvVars = project.GetProjectEnvVars(p, project.ProjectEnvVarParams{
			ApiUrl:        s.serverApiUrl,
			ServerUrl:     s.serverUrl,
			ServerVersion: s.serverVersion,
			ClientId:      telemetry.ClientId(ctx),
		}, telemetry.TelemetryEnabled(ctx))

		for k, v := range p.EnvVars {
			projectWithEnv.EnvVars[k] = v
		}

		var err error

		p = &projectWithEnv

		ws.Projects[i] = p
		err = s.workspaceStore.Save(ws)
		if err != nil {
			return nil, err
		}

		err = s.createProject(p, targetConfig, projectLogger)
		if err != nil {
			return nil, err
		}
	}

	wsLogger.Write([]byte("Workspace creation complete. Pending start...\n"))

	err = s.startWorkspace(ctx, ws, targetConfig, wsLogger)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *WorkspaceService) getCachedBuildForProject(p *project.Project) (*buildconfig.CachedBuild, error) {
	validStates := &[]build.BuildState{
		build.BuildState(build.BuildStatePublished),
	}

	build, err := s.buildService.Find(&build.Filter{
		States:        validStates,
		RepositoryUrl: &p.Repository.Url,
		Branch:        &p.Repository.Branch,
		EnvVars:       &p.EnvVars,
		BuildConfig:   p.BuildConfig,
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
