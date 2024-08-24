// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"

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
		Id:     req.Id,
		Name:   req.Name,
		Target: req.Target,
	}

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, w.Id)
	if err != nil {
		return nil, err
	}
	w.ApiKey = apiKey

	w.Projects = []*project.Project{}
	for _, p := range req.Projects {
		projectConfig := conversion.ToProjectConfig(p)

		isValidProjectName := regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`).MatchString
		if !isValidProjectName(projectConfig.Name) {
			return nil, ErrInvalidProjectName
		}

		if projectConfig.Repository != nil {
			projectConfig.Repository.Url = util.CleanUpRepositoryUrl(projectConfig.Repository.Url)
			if projectConfig.Repository.Sha == "" {
				sha, err := s.gitProviderService.GetLastCommitSha(projectConfig.Repository)
				if err != nil {
					return nil, err
				}
				projectConfig.Repository.Sha = sha
			}
		}

		if projectConfig.Image == "" {
			projectConfig.Image = s.defaultProjectImage
		}

		if projectConfig.User == "" {
			projectConfig.User = s.defaultProjectUser
		}
		log.Infoln(fmt.Sprintf("CreateWorkspace__identity 1: %s", p.Identity))
		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", w.Id, projectConfig.Name))
		if err != nil {
			return nil, err
		}
		project := &project.Project{
			ProjectConfig: *projectConfig,
			WorkspaceId:   w.Id,
			ApiKey:        apiKey,
			Target:        w.Target,
			Identity:      p.Identity,
		}
		w.Projects = append(w.Projects, project)
	}

	err = s.workspaceStore.Save(w)
	if err != nil {
		return nil, err
	}

	target, err := s.targetStore.Find(w.Target)
	if err != nil {
		return w, err
	}

	w, err = s.createWorkspace(ctx, w, target)

	if !telemetry.TelemetryEnabled(ctx) {
		return w, err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w, target)
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

func (s *WorkspaceService) createWorkspace(ctx context.Context, ws *workspace.Workspace, target *provider.ProviderTarget) (*workspace.Workspace, error) {
	wsLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, logs.LogSourceServer)
	defer wsLogger.Close()

	wsLogger.Write([]byte(fmt.Sprintf("Creating workspace %s (%s)\n", ws.Name, ws.Id)))

	ws.EnvVars = workspace.GetWorkspaceEnvVars(ws, workspace.WorkspaceEnvVarParams{
		ApiUrl:    s.serverApiUrl,
		ServerUrl: s.serverUrl,
		ClientId:  telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err := s.provisioner.CreateWorkspace(ws, target)
	if err != nil {
		return nil, err
	}

	for i, p := range ws.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(ws.Id, p.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		projectWithEnv := *p
		projectWithEnv.EnvVars = project.GetProjectEnvVars(p, project.ProjectEnvVarParams{
			ApiUrl:    s.serverApiUrl,
			ServerUrl: s.serverUrl,
			ClientId:  telemetry.ClientId(ctx),
		}, telemetry.TelemetryEnabled(ctx))

		for k, v := range p.EnvVars {
			projectWithEnv.EnvVars[k] = v
		}
		var gitProviderConfig *gitprovider.GitProviderConfig
		if p.Identity == "" {
			gc, err := s.gitProviderService.GetConfigForUrl(p.Repository.Url)
			if err != nil && !gitprovider.IsGitProviderNotFound(err) {
				return nil, err
			}
			p.Identity = gc.TokenIdentity
			gitProviderConfig = gc
		} else {
			gc, err := s.gitProviderService.GetConfig("", p.Identity)
			if err != nil {
				return nil, err
			}
			gitProviderConfig = gc
		}
		wsLogger.Write([]byte(fmt.Sprintf("createWorkspace__identity 2: %s", p.Identity)))

		var err error

		p = &projectWithEnv

		ws.Projects[i] = p
		err = s.workspaceStore.Save(ws)
		if err != nil {
			return nil, err
		}
		err = s.createProject(p, gitProviderConfig, target, projectLogger)
		if err != nil {
			return nil, err
		}
	}

	wsLogger.Write([]byte("Workspace creation complete. Pending start...\n"))

	err = s.startWorkspace(ctx, ws, target, wsLogger)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *WorkspaceService) createProject(p *project.Project, gitProviderConfig *gitprovider.GitProviderConfig, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", p.Name)))

	cr, err := s.containerRegistryService.FindByImageName(p.Image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	err = s.provisioner.CreateProject(p, target, cr, gitProviderConfig)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s created\n", p.Name)))

	return nil
}
