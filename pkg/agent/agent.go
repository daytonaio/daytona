// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	log "github.com/sirupsen/logrus"
)

func (a *Agent) Start() error {
	log.Info("Starting Daytona Agent")

	project, err := a.getProject()
	if err != nil {
		return err
	}

	if project.Repository.Url == nil {
		return errors.New("repository url not found")
	}

	gitProvider, err := a.getGitProvider(*project.Repository.Url)
	if err != nil {
		return err
	}

	var authToken *string = nil
	if gitProvider != nil {
		authToken = gitProvider.Token
	}

	exists, err := a.Git.RepositoryExists(project)
	if err != nil {
		log.Error(fmt.Sprintf("failed to clone repository: %s", err))
	} else {
		if exists {
			log.Info("Repository already exists. Skipping clone...")
		} else {
			log.Info("Cloning repository...")
			err = a.Git.CloneRepository(project, authToken)
			if err != nil {
				log.Error(fmt.Sprintf("failed to clone repository: %s", err))
			} else {
				log.Info("Repository cloned")
			}
		}
	}

	var gitUser *serverapiclient.GitUser
	if gitProvider != nil {
		gitUser, err = a.getGitUser(*gitProvider.Id)
		if err != nil {
			log.Error(fmt.Sprintf("failed to get git user data: %s", err))
		}
	}

	err = a.Git.SetGitConfig(gitUser)
	if err != nil {
		log.Error(fmt.Sprintf("failed to set git config: %s", err))
	}

	go func() {
		err := a.Ssh.Start()
		if err != nil {
			log.Error(fmt.Sprintf("failed to start ssh server: %s", err))
		}
	}()

	return a.Tailscale.Start()
}

func (a *Agent) getProject() (*serverapiclient.Project, error) {
	workspace, err := server.GetWorkspace(a.Config.WorkspaceId)
	if err != nil {
		return nil, err
	}

	for _, project := range workspace.Projects {
		if *project.Name == a.Config.ProjectName {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}

func (a *Agent) getGitProvider(repoUrl string) (*serverapiclient.GitProvider, error) {
	ctx := context.Background()

	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	encodedUrl := url.QueryEscape(repoUrl)
	gitProvider, res, err := apiClient.GitProviderAPI.GetGitProviderForUrl(ctx, encodedUrl).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	if gitProvider != nil {
		return gitProvider, nil
	}

	return nil, nil
}

func (a *Agent) getGitUser(gitProviderId string) (*serverapiclient.GitUser, error) {
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	userData, res, err := apiClient.GitProviderAPI.GetGitUser(context.Background(), gitProviderId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	return userData, nil
}
