// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"
	"io"
	"regexp"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/containerregistry"
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

func (s *WorkspaceService) createBuild(project *workspace.Project, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig, logWriter io.Writer) (*workspace.Project, error) {
	if project.Build != nil {
		builder := s.builderFactory.Create(*project, cr, gc)

		lastBuildResult, err := builder.LoadBuildResults()
		if err != nil {
			return nil, err
		}
		if lastBuildResult == nil {
			err := builder.Prepare()
			if err != nil {
				return nil, err
			}

			builderPlugin := builder.GetBuilderPlugin()
			if builderPlugin != nil {
				buildResult, err := builderPlugin.Build()
				if err != nil {
					return nil, err
				}

				err = builderPlugin.Publish()
				if err != nil {
					return nil, err
				}

				err = builder.SaveBuildResults(*buildResult)
				if err != nil {
					return nil, err
				}

				/*
					err = (*targetProvider).PublishBuildArtifacts(project, ba)
					if err != nil {
						return err
					}
				*/

				//	todo: just for test
				lastBuildResult = buildResult
			}
		} else {
			logWriter.Write([]byte("Project build image found. Skipping build\n"))
		}

		if lastBuildResult != nil {
			project.Image = lastBuildResult.ImageName
			project.User = lastBuildResult.User
			project.Build = nil
			project.PostStartCommands = lastBuildResult.PostStartCommands
			project.PostCreateCommands = lastBuildResult.PostCreateCommands
		}
	}

	return project, nil
}

func (s *WorkspaceService) createProject(project *workspace.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

	cr, _ := s.containerRegistryStore.Find(project.GetImageServer())

	gc, _ := s.gitProviderService.GetConfigForUrl(project.Repository.Url)

	err := s.provisioner.CreateProject(project, target, cr, gc)
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

		cr, _ := s.containerRegistryStore.Find(project.GetImageServer())

		gc, _ := s.gitProviderService.GetConfigForUrl(project.Repository.Url)

		projectWithEnv := *project
		projectWithEnv.EnvVars = workspace.GetProjectEnvVars(project, s.serverApiUrl, s.serverUrl)

		for k, v := range project.EnvVars {
			projectWithEnv.EnvVars[k] = v
		}

		var err error

		project, err = s.createBuild(&projectWithEnv, cr, gc, projectLogger)
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
