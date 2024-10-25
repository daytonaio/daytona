// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

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
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/daytonaio/daytona/pkg/target/project/buildconfig"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func isValidTargetName(name string) bool {
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

func (s *TargetService) CreateTarget(ctx context.Context, req dto.CreateTargetDTO) (*target.Target, error) {
	_, err := s.targetStore.Find(req.Name)
	if err == nil {
		return nil, ErrTargetAlreadyExists
	}

	// Repo name is taken as the name for target by default
	if !isValidTargetName(req.Name) {
		return nil, ErrInvalidTargetName
	}

	target := &target.Target{
		Id:           req.Id,
		Name:         req.Name,
		TargetConfig: req.TargetConfig,
	}

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeTarget, target.Id)
	if err != nil {
		return nil, err
	}
	target.ApiKey = apiKey

	target.Projects = []*project.Project{}

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

		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", target.Id, p.Name))
		if err != nil {
			return nil, err
		}

		p.TargetId = target.Id
		p.ApiKey = apiKey
		p.TargetConfig = target.TargetConfig
		target.Projects = append(target.Projects, p)
	}

	err = s.targetStore.Save(target)
	if err != nil {
		return nil, err
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return target, err
	}

	target, err = s.createTarget(ctx, target, targetConfig)

	if !telemetry.TelemetryEnabled(ctx) {
		return target, err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, targetConfig)
	event := telemetry.ServerEventTargetCreated
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetCreateError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return target, err
}

func (s *TargetService) createProject(p *project.Project, targetConfig *provider.TargetConfig, logWriter io.Writer) error {
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

func (s *TargetService) createTarget(ctx context.Context, t *target.Target, targetConfig *provider.TargetConfig) (*target.Target, error) {
	targetLogger := s.loggerFactory.CreateTargetLogger(t.Id, logs.LogSourceServer)
	defer targetLogger.Close()

	targetLogger.Write([]byte(fmt.Sprintf("Creating target %s (%s)\n", t.Name, t.Id)))

	t.EnvVars = target.GetTargetEnvVars(t, target.TargetEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err := s.provisioner.CreateTarget(t, targetConfig)
	if err != nil {
		return nil, err
	}

	for i, p := range t.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(t.Id, p.Name, logs.LogSourceServer)
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

		t.Projects[i] = p
		err = s.targetStore.Save(t)
		if err != nil {
			return nil, err
		}

		err = s.createProject(p, targetConfig, projectLogger)
		if err != nil {
			return nil, err
		}
	}

	targetLogger.Write([]byte("Target creation complete. Pending start...\n"))

	err = s.startTarget(ctx, t, targetConfig, targetLogger)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TargetService) getCachedBuildForProject(p *project.Project) (*buildconfig.CachedBuild, error) {
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
