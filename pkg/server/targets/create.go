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
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/target/workspace/buildconfig"
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

	target.Workspaces = []*workspace.Workspace{}

	for _, workspaceDto := range req.Workspaces {
		w := conversion.CreateDtoToWorkspace(workspaceDto)

		isValidWorkspaceName := regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`).MatchString
		if !isValidWorkspaceName(w.Name) {
			return nil, ErrInvalidWorkspaceName
		}

		w.Repository.Url = util.CleanUpRepositoryUrl(w.Repository.Url)
		if w.GitProviderConfigId == nil || *w.GitProviderConfigId == "" {
			configs, err := s.gitProviderService.ListConfigsForUrl(w.Repository.Url)
			if err != nil {
				return nil, err
			}

			if len(configs) > 1 {
				return nil, errors.New("multiple git provider configs found for the repository url")
			}

			if len(configs) == 1 {
				w.GitProviderConfigId = &configs[0].Id
			}
		}

		if w.Repository.Sha == "" {
			sha, err := s.gitProviderService.GetLastCommitSha(w.Repository)
			if err != nil {
				return nil, err
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

		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, fmt.Sprintf("%s/%s", target.Id, w.Name))
		if err != nil {
			return nil, err
		}

		w.TargetId = target.Id
		w.ApiKey = apiKey
		w.TargetConfig = target.TargetConfig
		target.Workspaces = append(target.Workspaces, w)
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

func (s *TargetService) createWorkspace(w *workspace.Workspace, targetConfig *provider.TargetConfig, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating workspace %s\n", w.Name)))

	cr, err := s.containerRegistryService.FindByImageName(w.Image)
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

	err = s.provisioner.CreateWorkspace(w, targetConfig, cr, gc)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Workspace %s created\n", w.Name)))

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

	for i, w := range t.Workspaces {
		workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(t.Id, w.Name, logs.LogSourceServer)
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

		var err error

		w = &workspaceWithEnv

		t.Workspaces[i] = w
		err = s.targetStore.Save(t)
		if err != nil {
			return nil, err
		}

		err = s.createWorkspace(w, targetConfig, workspaceLogger)
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

func (s *TargetService) getCachedBuildForWorkspace(w *workspace.Workspace) (*buildconfig.CachedBuild, error) {
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
