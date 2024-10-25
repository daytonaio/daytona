// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
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
)

func NewTargetEventProps(ctx context.Context, target *target.Target, targetConfig *provider.TargetConfig) map[string]interface{} {
	props := map[string]interface{}{}

	sessionId := SessionId(ctx)
	serverId := ServerId(ctx)

	props["session_id"] = sessionId
	props["server_id"] = serverId

	if target != nil {
		props["target_id"] = target.Id
		props["target_n_projects"] = len(target.Projects)
		publicRepos := []string{}
		publicImages := []string{}
		builders := map[string]int{}

		for _, project := range target.Projects {
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

		props["target_public_repos"] = publicRepos
		props["target_public_images"] = publicImages
		props["target_builders"] = builders
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
