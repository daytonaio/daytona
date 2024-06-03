// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	agent_config "github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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
	err := a.setDefaultConfig()
	if err != nil {
		return err
	}

	project, err := a.getProject()
	if err != nil {
		return err
	}

	// Ignoring error because we don't want to fail if the git provider is not found
	gitProvider, _ := a.getGitProvider(project.Repository.Url)

	var auth *http.BasicAuth
	if gitProvider != nil {
		auth = &http.BasicAuth{}
		if gitProvider.Username != nil {
			auth.Username = *gitProvider.Username
		}
		if gitProvider.Token != nil {
			auth.Password = *gitProvider.Token
		}
	}

	exists, err := a.Git.RepositoryExists(project)
	if err != nil {
		log.Error(fmt.Sprintf("failed to clone repository: %s", err))
	} else {
		if exists {
			log.Info("Repository already exists. Skipping clone...")
		} else {
			if stat, err := os.Stat(a.Config.ProjectDir); err == nil {
				ownerUid := stat.Sys().(*syscall.Stat_t).Uid
				if ownerUid != uint32(os.Getuid()) {
					chownCmd := exec.Command("sudo", "chown", "-R", fmt.Sprintf("%s:%s", project.User, project.User), a.Config.ProjectDir)
					err = chownCmd.Run()
					if err != nil {
						log.Error(err)
					}
				}
			}

			log.Info("Cloning repository...")
			err = a.Git.CloneRepository(project, auth)
			if err != nil {
				log.Error(fmt.Sprintf("failed to clone repository: %s", err))
			} else {
				log.Info("Repository cloned")
			}
		}
	}

	var gitUser *gitprovider.GitUser
	if gitProvider != nil {
		user, err := a.getGitUser(*gitProvider.Id)
		if err != nil {
			log.Error(fmt.Sprintf("failed to get git user data: %s", err))
		}
		gitUser = &gitprovider.GitUser{
			Email:    *user.Email,
			Name:     *user.Name,
			Id:       *user.Id,
			Username: *user.Username,
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

	err = a.runPostCreateCommands(project)
	if err != nil {
		log.Error(fmt.Sprintf("failed to run post create commands: %s", err))
		log.Error("skipping post start commands...")
	} else {
		a.runPostStartCommands(project)
	}

	return nil
}

func (a *Agent) getProject() (*workspace.Project, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, a.Config.WorkspaceId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	for _, project := range workspace.Projects {
		if *project.Name == a.Config.ProjectName {
			return conversion.ToProject(&project), nil
		}
	}

	return nil, errors.New("project not found")
}

func (a *Agent) getGitProvider(repoUrl string) (*apiclient.GitProvider, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return nil, err
	}

	encodedUrl := url.QueryEscape(repoUrl)
	gitProvider, res, err := apiClient.GitProviderAPI.GetGitProviderForUrl(ctx, encodedUrl).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if gitProvider != nil {
		return gitProvider, nil
	}

	return nil, nil
}

func (a *Agent) getGitUser(gitProviderId string) (*apiclient.GitUser, error) {
	apiClient, err := apiclient_util.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return nil, err
	}

	userData, res, err := apiClient.GitProviderAPI.GetGitUser(context.Background(), gitProviderId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
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

	return config.Save()
}

func (a *Agent) runPostCreateCommands(project *workspace.Project) error {
	if _, err := os.Stat(a.PostCreateLockFilePath); err == nil {
		log.Info("Post create commands already ran. Skipping...")
		return nil
	}

	if len(project.PostCreateCommands) == 0 {
		log.Info("No post create commands to run")
		return nil
	}

	log.Info("Running post create commands...")

	var errCh = make(chan error)

	for _, command := range project.PostCreateCommands {
		go func() {
			log.Info("Running command: " + command)
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = a.Config.ProjectDir
			cmd.Stdout = io.MultiWriter(a.LogWriter, os.Stdout)
			cmd.Stderr = io.MultiWriter(a.LogWriter, os.Stderr)

			err := cmd.Run()
			if err != nil {
				log.Error(fmt.Sprintf("command '%s' failed: %v", command, err))
			}
			errCh <- err
		}()
	}

	resultErr := errors.New("errors: ")
	for range project.PostCreateCommands {
		err := <-errCh
		if err != nil {
			resultErr = fmt.Errorf("%w %v", resultErr, err)
		}
	}

	if resultErr.Error() != "errors: " {
		return resultErr
	}

	return os.WriteFile(a.PostCreateLockFilePath, []byte{}, 0644)
}

func (a *Agent) runPostStartCommands(project *workspace.Project) {
	log.Info("Running post start commands...")

	for _, command := range project.PostStartCommands {
		go func() {
			log.Info("Running command: " + command)
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = a.Config.ProjectDir
			cmd.Stdout = io.MultiWriter(a.LogWriter, os.Stdout)
			cmd.Stderr = io.MultiWriter(a.LogWriter, os.Stderr)

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
	apiClient, err := apiclient_util.GetAgentApiClient(a.Config.Server.ApiUrl, a.Config.Server.ApiKey)
	if err != nil {
		return err
	}

	gitStatus, err := a.Git.GetGitStatus()
	if err != nil {
		return err
	}

	uptime := a.uptime()
	res, err := apiClient.WorkspaceAPI.SetProjectState(context.Background(), a.Config.WorkspaceId, a.Config.ProjectName).SetState(apiclient.SetProjectState{
		Uptime:    &uptime,
		GitStatus: conversion.ToGitStatusDTO(gitStatus),
	}).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	return nil
}
