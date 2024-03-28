// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

func CreateWorkspace(createWorkspaceDto dto.CreateWorkspace) (*types.Workspace, error) {
	_, err := db.FindWorkspaceByName(createWorkspaceDto.Name)
	if err == nil {
		return nil, errors.New("workspace already exists")
	}

	w, err := newWorkspace(createWorkspaceDto)
	if err != nil {
		return nil, err
	}

	err = db.SaveWorkspace(w)
	if err != nil {
		return nil, err
	}

	err = provisioner.CreateWorkspace(w)
	if err != nil {
		return nil, err
	}

	err = provisioner.StartWorkspace(w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func newWorkspace(createWorkspaceDto dto.CreateWorkspace) (*types.Workspace, error) {
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(createWorkspaceDto.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	_, err := targets.GetTarget(createWorkspaceDto.Target)
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

	return w, nil
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
