// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/git"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	log "github.com/sirupsen/logrus"
)

func Start() error {
	log.Info("Starting Daytona Agent")

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	project, err := getProject(c)
	if err != nil {
		return err
	}

	if project.Repository.Url == nil {
		return errors.New("repository url not found")
	}

	gitProvider, err := getGitProvider(*project.Repository.Url)
	if err != nil {
		return err
	}

	var authToken *string = nil
	if gitProvider != nil {
		authToken = gitProvider.Token
	}

	if _, err := os.Stat(c.ProjectDir); os.IsNotExist(err) {
		log.Info("Cloning repository...")
		err = git.CloneRepository(c, project, authToken)
		if err != nil {
			log.Error(fmt.Sprintf("failed to clone repository: %s", err))
		} else {
			log.Info("Repository cloned")
		}
	} else {
		log.Info("Repository already exists. Skipping clone...")
	}

	var gitUserData *serverapiclient.GitUserData
	if gitProvider != nil {
		gitUserData, err = getGitUserData(*gitProvider.Id)
		if err != nil {
			log.Error(fmt.Sprintf("failed to get git user data: %s", err))
		}
	}

	err = git.SetGitConfig(gitUserData)
	if err != nil {
		log.Error(fmt.Sprintf("failed to set git config: %s", err))
	}

	go func() {
		ssh.Start()
	}()

	return tailscale.Start(c)
}

func getProject(c *config.Config) (*serverapiclient.Project, error) {
	workspace, err := server.GetWorkspace(c.WorkspaceId)
	if err != nil {
		return nil, err
	}

	for _, project := range workspace.Projects {
		if *project.Name == c.ProjectName {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}

func getGitProvider(repoUrl string) (*serverapiclient.GitProvider, error) {
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	gitProvider := gitprovider.GetGitProviderFromHost(repoUrl, serverConfig.GitProviders)
	if gitProvider != nil {
		return gitProvider, nil
	}

	return nil, nil
}

func getGitUserData(gitProviderId string) (*serverapiclient.GitUserData, error) {
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return nil, err
	}

	userData, res, err := apiClient.GitProviderAPI.GetGitUserData(context.Background(), gitProviderId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	return userData, nil
}
