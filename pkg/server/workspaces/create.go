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

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

func CreateWorkspace(createWorkspaceDto dto.CreateWorkspace) (*types.Workspace, error) {
	_, err := db.FindWorkspaceByName(createWorkspaceDto.Name)
	if err == nil {
		return nil, errors.New("workspace already exists")
	}

	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(createWorkspaceDto.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	t, err := targets.GetTarget(createWorkspaceDto.Target)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{
		Id:     uuid.NewString(),
		Name:   createWorkspaceDto.Name,
		Target: createWorkspaceDto.Target,
	}

	w.Projects = []*types.Project{}
	serverConfig, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	userGitProviders := serverConfig.GitProviders

	for _, repo := range createWorkspaceDto.Repositories {
		providerId := getGitProviderIdFromUrl(repo.Url)
		gitProvider := gitprovider.GetGitProvider(providerId, userGitProviders)

		if gitProvider != nil {
			gitUser, err := gitProvider.GetUser()
			if err != nil {
				return nil, err
			}
			repo.GitUser = &types.GitUser{
				Name:  gitUser.Name,
				Email: gitUser.Email,
			}
		}

		projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
		projectName := projectNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repo.Url)), ".git"), "-")

		apiKey, err := auth.GenerateApiKey(types.ApiKeyTypeProject, fmt.Sprintf("%s/%s", w.Id, projectName))
		if err != nil {
			return nil, err
		}

		project := &types.Project{
			Name:        projectName,
			Repository:  &repo,
			WorkspaceId: w.Id,
			ApiKey:      apiKey,
			Target:      createWorkspaceDto.Target,
		}
		w.Projects = append(w.Projects, project)
	}

	err = db.SaveWorkspace(w)
	if err != nil {
		return nil, err
	}

	err = provisioner.CreateWorkspace(w, t)
	if err != nil {
		return nil, err
	}

	err = provisioner.StartWorkspace(w, t)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func createProject(project *types.Project, c *types.ServerConfig, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

	projectToCreate := *project
	projectToCreate.EnvVars = getProjectEnvVars(project, c)

	err := provisioner.CreateProject(&projectToCreate, target)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s created\n", project.Name)))

	return nil
}

func createWorkspace(workspace *types.Workspace) (*types.Workspace, error) {
	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return workspace, err
	}

	logsDir, err := config.GetWorkspaceLogsDir()
	if err != nil {
		return workspace, err
	}

	c, err := config.GetConfig()
	if err != nil {
		return workspace, err
	}

	workspaceLogger := logger.GetWorkspaceLogger(logsDir, workspace.Id)
	defer workspaceLogger.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, workspaceLogger)

	wsLogWriter.Write([]byte("Creating workspace\n"))

	err = provisioner.CreateWorkspace(workspace, target)
	if err != nil {
		return nil, err
	}

	for _, project := range workspace.Projects {
		projectLogger := logger.GetProjectLogger(logsDir, workspace.Id, project.Name)
		defer projectLogger.Close()

		projectLogWriter := io.MultiWriter(wsLogWriter, projectLogger)
		err := createProject(project, c, target, projectLogWriter)
		if err != nil {
			return nil, err
		}
	}

	wsLogWriter.Write([]byte("Workspace creation complete. Pending start...\n"))

	err = startWorkspace(workspace, target, c, logsDir, wsLogWriter)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func getProjectEnvVars(project *types.Project, config *types.ServerConfig) map[string]string {
	envVars := map[string]string{
		"DAYTONA_WS_ID":                     project.WorkspaceId,
		"DAYTONA_WS_PROJECT_NAME":           project.Name,
		"DAYTONA_WS_PROJECT_REPOSITORY_URL": project.Repository.Url,
		"DAYTONA_SERVER_API_KEY":            project.ApiKey,
		"DAYTONA_SERVER_VERSION":            internal.Version,
		"DAYTONA_SERVER_URL":                util.GetFrpcServerUrl(config),
		"DAYTONA_SERVER_API_URL":            util.GetFrpcApiUrl(config),
	}

	return envVars
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
