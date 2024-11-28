// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
)

type ServerEvent string

const (
	ServerEventApiRequestStarted ServerEvent = "server_api_request_started"
	ServerEventApiResponseSent   ServerEvent = "server_api_response_sent"
	ServerEventPurgeStarted      ServerEvent = "server_purge_started"
	ServerEventPurgeCompleted    ServerEvent = "server_purge_completed"
	ServerEventPurgeError        ServerEvent = "server_purge_error"

	// Target events
	ServerEventTargetCreated      ServerEvent = "server_target_created"
	ServerEventTargetDestroyed    ServerEvent = "server_target_destroyed"
	ServerEventTargetStarted      ServerEvent = "server_target_started"
	ServerEventTargetStopped      ServerEvent = "server_target_stopped"
	ServerEventTargetCreateError  ServerEvent = "server_target_created_error"
	ServerEventTargetDestroyError ServerEvent = "server_target_destroyed_error"
	ServerEventTargetStartError   ServerEvent = "server_target_started_error"
	ServerEventTargetStopError    ServerEvent = "server_target_stopped_error"

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

func NewTargetEventProps(ctx context.Context, target *models.Target) map[string]interface{} {
	props := map[string]interface{}{}

	sessionId := SessionId(ctx)
	serverId := ServerId(ctx)

	props["session_id"] = sessionId
	props["server_id"] = serverId

	if target != nil {
		props["target_id"] = target.Id
		props["target_name"] = target.Name
		props["target_provider"] = target.TargetConfig.ProviderInfo.Name
		props["target_provider_version"] = target.TargetConfig.ProviderInfo.Version
	}

	return props
}

func NewWorkspaceEventProps(ctx context.Context, workspace *models.Workspace) map[string]interface{} {
	props := map[string]interface{}{}

	if workspace == nil {
		return props
	}

	props["workspace_provider"] = workspace.Target.TargetConfig.ProviderInfo.Name
	props["workspace_provider_version"] = workspace.Target.TargetConfig.ProviderInfo.Version

	if isImagePublic(workspace.Image) {
		props["workspace_image"] = workspace.Image
	}
	if workspace.Repository != nil && isPublic(workspace.Repository.Url) {
		props["workspace_repository"] = workspace.Repository.Url
	}

	if workspace.BuildConfig == nil {
		props["workspace_builder"] = "none"
	} else if workspace.BuildConfig.Devcontainer != nil {
		props["workspace_builder"] = "devcontainer"
	} else {
		props["workspace_builder"] = "automatic"
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
