// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type ServerEvent string

const (
	ServerEventApiRequestStarted ServerEvent = "server_api_request_started"
	ServerEventApiResponseSent   ServerEvent = "server_api_response_sent"
	ServerEventPurgeStarted      ServerEvent = "server_purge_started"
	ServerEventPurgeCompleted    ServerEvent = "server_purge_completed"
	ServerEventPurgeError        ServerEvent = "server_purge_error"

	// Workspace events
	ServerEventWorkspaceCreated      ServerEvent = "server_workspace_created"
	ServerEventWorkspaceDestroyed    ServerEvent = "server_workspace_destroyed"
	ServerEventWorkspaceStarted      ServerEvent = "server_workspace_started"
	ServerEventWorkspaceStopped      ServerEvent = "server_workspace_stopped"
	ServerEventWorkspaceCreateError  ServerEvent = "server_workspace_created_error"
	ServerEventWorkspaceDestroyError ServerEvent = "server_workspace_destroyed_error"
	ServerEventWorkspaceStartError   ServerEvent = "server_workspace_started_error"
	ServerEventWorkspaceStopError    ServerEvent = "server_workspace_stopped_error"
)

func NewWorkspaceEventProps(ctx context.Context, workspace *workspace.Workspace, targetConfig *provider.TargetConfig) map[string]interface{} {
	props := map[string]interface{}{}

	sessionId := SessionId(ctx)
	serverId := ServerId(ctx)

	props["session_id"] = sessionId
	props["server_id"] = serverId

	if workspace != nil {
		props["workspace_id"] = workspace.Id
		props["workspace_n_projects"] = len(workspace.Projects)
		publicRepos := []string{}
		publicImages := []string{}
		builders := map[string]int{}

		for _, project := range workspace.Projects {
			if isImagePublic(project.Image) {
				publicImages = append(publicImages, project.Image)
			}
			if project.Repository != nil && isPublic(project.Repository.Url) {
				publicRepos = append(publicRepos, project.Repository.Url)
			}
			if project.BuildConfig == nil {
				builders["none"]++
			} else if project.BuildConfig.Devcontainer != nil {
				builders["devcontainer"]++
			} else {
				builders["automatic"]++
			}
		}

		props["workspace_public_repos"] = publicRepos
		props["workspace_public_images"] = publicImages
		props["workspace_builders"] = builders
	}

	if targetConfig != nil {
		props["target_name"] = targetConfig.Name
		props["target_provider"] = targetConfig.ProviderInfo.Name
		props["target_provider_version"] = targetConfig.ProviderInfo.Version
	}

	return props
}

func isImagePublic(imageName string) bool {
	if strings.Count(imageName, "/") < 2 {
		if strings.Count(imageName, "/") == 0 {
			return isPublic("https://hub.docker.com/_/" + imageName)
		}

		return isPublic("https://hub.docker.com/r/" + imageName)
	}

	return isPublic(imageName)
}

func isPublic(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	_, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	cancel()
	return err == nil
}
