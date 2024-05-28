// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"
	"io"
	"regexp"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/gitprovider"
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

		postStartCommands := s.defaultProjectPostStartCommands
		if project.PostStartCommands != nil {
			postStartCommands = *project.PostStartCommands
		}

		p := &workspace.Project{
			Name:              project.Name,
			Image:             projectImage,
			User:              projectUser,
			Build:             project.Build,
			PostStartCommands: postStartCommands,
			Repository:        project.Source.Repository,
			WorkspaceId:       w.Id,
			ApiKey:            apiKey,
			Target:            w.Target,
			EnvVars:           project.EnvVars,
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
	if project.Build != nil {
		lastBuildResult, err := s.builderFactory.CheckExistingBuild(*project)
		if err != nil {
			return nil, err
		}
		if lastBuildResult != nil {
			project.Image = lastBuildResult.ImageName
			project.User = lastBuildResult.User
			project.Build = nil
			project.PostStartCommands = lastBuildResult.PostStartCommands
			project.PostCreateCommands = lastBuildResult.PostCreateCommands
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
			cleanupErr := builder.CleanUp()
			if cleanupErr != nil {
				logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())))
			}

			return nil, err
		}

		err = builder.Publish()
		if err != nil {
			cleanupErr := builder.CleanUp()
			if cleanupErr != nil {
				logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())))
			}
			return nil, err
		}

		err = builder.SaveBuildResults(*buildResult)
		if err != nil {
			cleanupErr := builder.CleanUp()
			if cleanupErr != nil {
				logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())))
			}
			return nil, err
		}

		err = builder.CleanUp()
		if err != nil {
			return nil, err
		}

		project.Image = buildResult.ImageName
		project.User = buildResult.User
		project.Build = nil
		project.PostStartCommands = buildResult.PostStartCommands
		project.PostCreateCommands = buildResult.PostCreateCommands

		return project, nil
	}

	return project, nil
}

func (s *WorkspaceService) createProject(project *workspace.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

	cr, _ := s.containerRegistryStore.Find(project.GetImageServer())

	err := s.provisioner.CreateProject(project, target, cr)
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

	wsLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id)
	wsLogger.Write([]byte("Creating workspace\n"))

	err = s.provisioner.CreateWorkspace(ws, target)
	if err != nil {
		return nil, err
	}

	for i, project := range ws.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(ws.Id, project.Name)
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
