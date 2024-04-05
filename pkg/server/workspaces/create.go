// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/google/uuid"
)

func (s *WorkspaceService) CreateWorkspace(name string, targetId string, repositories []gitprovider.GitRepository) (*workspace.Workspace, error) {
	_, err := s.workspaceStore.Find(name)
	if err == nil {
		return nil, errors.New("workspace already exists")
	}

	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	providerName, targetName, err := s.parseTargetId(targetId)
	if err != nil {
		return nil, err
	}

	t, err := s.targetStore.Find(providerName, targetName)
	if err != nil {
		return nil, err
	}

	w := &workspace.Workspace{
		Id:     uuid.NewString(),
		Name:   name,
		Target: targetId,
	}

	w.Projects = []*workspace.Project{}

	for _, repo := range repositories {
		projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
		projectName := projectNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repo.Url)), ".git"), "-")

		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", w.Id, projectName))
		if err != nil {
			return nil, err
		}

		project := &workspace.Project{
			Name:        projectName,
			Repository:  &repo,
			WorkspaceId: w.Id,
			ApiKey:      apiKey,
			Target:      targetId,
		}
		w.Projects = append(w.Projects, project)
	}

	err = s.workspaceStore.Save(w)
	if err != nil {
		return nil, err
	}

	err = s.provisioner.CreateWorkspace(w, t)
	if err != nil {
		return nil, err
	}

	err = s.provisioner.StartWorkspace(w, t)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *WorkspaceService) createProject(project *workspace.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

	projectToCreate := *project
	projectToCreate.EnvVars = workspace.GetProjectEnvVars(project, s.serverApiUrl, s.serverUrl)

	err := s.provisioner.CreateProject(&projectToCreate, target)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s created\n", project.Name)))

	return nil
}

func (s *WorkspaceService) createWorkspace(workspace *workspace.Workspace) (*workspace.Workspace, error) {
	providerName, targetName, err := s.parseTargetId(workspace.Target)
	if err != nil {
		return workspace, err
	}

	target, err := s.targetStore.Find(providerName, targetName)
	if err != nil {
		return workspace, err
	}

	wsLogger := s.newWorkspaceLogger(workspace.Id)
	wsLogger.Write([]byte("Creating workspace\n"))

	err = s.provisioner.CreateWorkspace(workspace, target)
	if err != nil {
		return nil, err
	}

	for _, project := range workspace.Projects {
		projectLogger := s.newProjectLogger(workspace.Id, project.Name)
		defer projectLogger.Close()

		projectLogWriter := io.MultiWriter(wsLogger, projectLogger)
		err := s.createProject(project, target, projectLogWriter)
		if err != nil {
			return nil, err
		}
	}

	wsLogger.Write([]byte("Workspace creation complete. Pending start...\n"))

	err = s.startWorkspace(workspace, target, wsLogger)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func getGitProviderIdFromUrl(url string) string {
	if strings.Contains(url, "github.com") {
		return "github"
	} else if strings.Contains(url, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(url, "bitbucket.org") {
		return "bitbucket"
	} else if strings.Contains(url, "codeberg.org") {
		return "codeberg"
	} else {
		return ""
	}
}
