// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"
	"io"
	"regexp"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/builder"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func (s *WorkspaceService) CreateWorkspace(req dto.CreateWorkspaceRequest) (*workspace.Workspace, error) {
	_, err := s.workspaceStore.Find(req.Name)
	if err == nil {
		return nil, ErrWorkspaceAlreadyExists
	}

	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(req.Name) {
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

	w.Projects = []*workspace.Project{}

	for _, project := range req.Projects {
		isValidProjectName := regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`).MatchString
		if !isValidProjectName(project.Name) {
			return nil, ErrInvalidProjectName
		}

		if project.Source.Repository != nil && project.Source.Repository.Sha == "" {
			sha, err := s.gitProviderService.GetLastCommitSha(project.Source.Repository)
			if err != nil {
				return nil, err
			}
			project.Source.Repository.Sha = sha
		}

		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", w.Id, project.Name))
		if err != nil {
			return nil, err
		}

		projectImage := s.defaultProjectImage
		if project.Image != nil {
			projectImage = *project.Image
		}

		projectUser := s.defaultProjectUser
		if project.User != nil {
			projectUser = *project.User
		}

		p := &workspace.Project{
			Name:        project.Name,
			Image:       projectImage,
			User:        projectUser,
			Build:       project.Build,
			Repository:  project.Source.Repository,
			WorkspaceId: w.Id,
			ApiKey:      apiKey,
			Target:      w.Target,
			EnvVars:     project.EnvVars,
		}
		w.Projects = append(w.Projects, p)
	}

	err = s.workspaceStore.Save(w)
	if err != nil {
		return nil, err
	}

	return s.createWorkspace(w)
}

func (s *WorkspaceService) createBuild(project *workspace.Project, gc *gitprovider.GitProviderConfig, logWriter io.Writer) (*workspace.Project, error) {
	// FIXME: skip build completely for now
	return project, nil

	if project.Build != nil { // nolint:govet
		lastBuildResult, err := s.builderFactory.CheckExistingBuild(*project)
		if err != nil {
			return nil, err
		}
		if lastBuildResult != nil {
			project.Image = lastBuildResult.ImageName
			project.User = lastBuildResult.User
			return project, nil
		}

		builder, err := s.builderFactory.Create(*project, gc)
		if err != nil {
			return nil, err
		}

		if builder == nil {
			return project, nil
		}

		buildResult, err := builder.Build()
		if err != nil {
			s.handleBuildError(project, builder, logWriter, err)
			return project, nil
		}

		err = builder.Publish()
		if err != nil {
			s.handleBuildError(project, builder, logWriter, err)
			return project, nil
		}

		err = builder.SaveBuildResults(*buildResult)
		if err != nil {
			s.handleBuildError(project, builder, logWriter, err)
			return project, nil
		}

		err = builder.CleanUp()
		if err != nil {
			logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", err.Error())))
		}

		project.Image = buildResult.ImageName
		project.User = buildResult.User

		return project, nil
	}

	return project, nil
}

func (s *WorkspaceService) createProject(project *workspace.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

	cr, err := s.containerRegistryService.FindByImageName(project.Image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	gc, err := s.gitProviderService.GetConfigForUrl(project.Repository.Url)
	if err != nil && !gitprovider.IsGitProviderNotFound(err) {
		return err
	}

	err = s.provisioner.CreateProject(project, target, cr, gc)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s created\n", project.Name)))

	return nil
}

func (s *WorkspaceService) createWorkspace(ws *workspace.Workspace) (*workspace.Workspace, error) {
	target, err := s.targetStore.Find(ws.Target)
	if err != nil {
		return ws, err
	}

	wsLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, logs.LogSourceServer)
	defer wsLogger.Close()

	wsLogger.Write([]byte(fmt.Sprintf("Creating workspace %s (%s)\n", ws.Name, ws.Id)))

	err = s.provisioner.CreateWorkspace(ws, target)
	if err != nil {
		return nil, err
	}

	for i, project := range ws.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(ws.Id, project.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		gc, _ := s.gitProviderService.GetConfigForUrl(project.Repository.Url)

		projectWithEnv := *project
		projectWithEnv.EnvVars = workspace.GetProjectEnvVars(project, s.serverApiUrl, s.serverUrl)

		for k, v := range project.EnvVars {
			projectWithEnv.EnvVars[k] = v
		}

		var err error

		project, err = s.createBuild(&projectWithEnv, gc, projectLogger)
		if err != nil {
			return nil, err
		}

		ws.Projects[i] = project
		err = s.workspaceStore.Save(ws)
		if err != nil {
			return nil, err
		}

		err = s.createProject(project, target, projectLogger)
		if err != nil {
			return nil, err
		}
	}

	wsLogger.Write([]byte("Workspace creation complete. Pending start...\n"))

	err = s.startWorkspace(ws, target, wsLogger)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *WorkspaceService) handleBuildError(project *workspace.Project, builder builder.IBuilder, logWriter io.Writer, err error) {
	logWriter.Write([]byte("################################################\n"))
	logWriter.Write([]byte(fmt.Sprintf("#### BUILD FAILED FOR PROJECT %s: %s\n", project.Name, err.Error())))
	logWriter.Write([]byte("################################################\n"))

	cleanupErr := builder.CleanUp()
	if cleanupErr != nil {
		logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())))
	}

	logWriter.Write([]byte("Creating project with default image\n"))
	project.Image = s.defaultProjectImage
	project.User = s.defaultProjectUser
}
