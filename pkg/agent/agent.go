// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	agent_config "github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	log "github.com/sirupsen/logrus"
)

func (a *Agent) Start() error {
	a.initLogs()

	log.Info("Starting Daytona Agent")

	a.startTime = time.Now()

	if a.Config.Mode == agent_config.ModeProject {
		err := a.startProjectMode()
		if err != nil {
			return err
		}
	}

	go func() {
		err := a.Ssh.Start()
		if err != nil {
			log.Error(fmt.Sprintf("failed to start ssh server: %s", err))
		}
	}()

	return a.Tailscale.Start()
}

func (a *Agent) startProjectMode() error {
	project, err := a.getProject()
	if err != nil {
		return err
	}
	a.project = project

	err = a.setDefaultConfig()
	if err != nil {
		return err
	}

	if a.project.Repository.Url == nil {
		return errors.New("repository url not found")
	}

	// Ignoring error because we don't want to fail if the git provider is not found
	gitProvider, _ := a.getGitProvider()

	var authToken *string = nil
	if gitProvider != nil {
		authToken = gitProvider.Token
	}

	exists, err := a.Git.RepositoryExists(a.project)
	if err != nil {
		log.Error(fmt.Sprintf("failed to clone repository: %s", err))
	} else {
		if exists {
			log.Info("Repository already exists. Skipping clone...")
		} else {
			log.Info("Cloning repository...")
			err = a.Git.CloneRepository(a.project, authToken)
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
		for {
			err := a.updateProjectState()
			if err != nil {
				log.Error(fmt.Sprintf("failed to update project state: %s", err))
			}

			time.Sleep(2 * time.Second)
		}
	}()

	a.runPostStartCommands()

	return nil
}

func (a *Agent) getProject() (*serverapiclient.Project, error) {
	ctx := context.Background()

	if a.Config == nil {
		config, err := agent_config.GetConfig(agent_config.ModeProject)
		if err != nil {
			return nil, err
		}
		a.Config = config
	}

	apiClient, err := server.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, a.Config.WorkspaceId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	for _, project := range workspace.Projects {
		if *project.Name == a.Config.ProjectName {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}

func (a *Agent) getGitProvider() (*serverapiclient.GitProvider, error) {
	ctx := context.Background()

	apiClient, err := server.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return nil, err
	}

	encodedUrl := url.QueryEscape(*a.project.Repository.Url)
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
	apiClient, err := server.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return nil, err
	}

	userData, res, err := apiClient.GitProviderAPI.GetGitUser(context.Background(), gitProviderId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	return userData, nil
}

func (a *Agent) setDefaultConfig() error {
	existingConfig, err := config.GetConfig()
	if err != nil && !config.IsNotExist(err) {
		return err
	}

	if existingConfig != nil {
		for _, profile := range existingConfig.Profiles {
			if profile.Id == "default" {
				return nil
			}
		}
	}

	config := &config.Config{
		ActiveProfileId: "default",
		DefaultIdeId:    "vscode",
		Profiles: []config.Profile{
			{
				Id:   "default",
				Name: "default",
				Api: config.ServerApi{
					Url: a.Config.Server.ApiUrl,
					Key: a.Config.Server.ApiKey,
				},
			},
		},
	}

	err = config.Save()
	if err != nil {
		return err
	}

	//	as the agent is running as root, we need to move the config to the user's home directory
	user, err := user.Lookup(*a.project.User)
	if err != nil {
		return err
	}

	rootConfigPath := "/root/.config/daytona"
	userConfigPath := user.HomeDir + "/.config/daytona"

	err = os.MkdirAll(user.HomeDir+"/.config", 0755)
	if err != nil {
		return err
	}

	err = os.Rename(rootConfigPath, userConfigPath)
	if err != nil {
		return err
	}

	return nil

}

func (a *Agent) runPostStartCommands() {
	log.Info("Running post start commands...")

	for _, command := range a.project.PostStartCommands {
		go func() {
			log.Info("Running command: " + command)
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = a.Config.ProjectDir
			cmd.Stdout = a.LogWriter
			cmd.Stderr = a.LogWriter

			err := cmd.Run()
			if err != nil {
				log.Error(fmt.Sprintf("command '%s' failed: %v", command, err))
			}
		}()
	}
}

// Agent uptime in seconds
func (a *Agent) uptime() int32 {
	return int32(time.Since(a.startTime).Seconds())
}

func (a *Agent) updateProjectState() error {
	apiClient, err := server.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return err
	}

	uptime := a.uptime()
	res, err := apiClient.WorkspaceAPI.SetProjectState(context.Background(), a.Config.WorkspaceId, a.Config.ProjectName).SetState(serverapiclient.SetProjectState{
		Uptime: &uptime,
	}).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	return nil
}
