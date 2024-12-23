// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "github.com/daytonaio/daytona/pkg/models"

type GitProviderConfigEventName string

const (
	GitProviderConfigEventLifecycleSaved               GitProviderConfigEventName = "git_provider_config_lifecycle_saved"
	GitProviderConfigEventLifecycleSaveFailed          GitProviderConfigEventName = "git_provider_config_lifecycle_save_failed"
	GitProviderConfigEventLifecycleDeleted             GitProviderConfigEventName = "git_provider_config_lifecycle_deleted"
	GitProviderConfigEventLifecycleDeletionFailed      GitProviderConfigEventName = "git_provider_config_lifecycle_deletion_failed"
	GitProviderConfigEventLifecycleForceDeleted        GitProviderConfigEventName = "git_provider_config_lifecycle_force_deleted"
	GitProviderConfigEventLifecycleForceDeletionFailed GitProviderConfigEventName = "git_provider_config_lifecycle_force_deletion_failed"
)

type GitProviderConfigEvent struct {
	GitProviderConfig *models.GitProviderConfig
	AbstractEvent
}

func NewGitProviderConfigEvent(name GitProviderConfigEventName, b *models.GitProviderConfig, err error, extras map[string]interface{}) Event {
	return GitProviderConfigEvent{
		GitProviderConfig: b,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e GitProviderConfigEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.GitProviderConfig != nil {
		props["provider_id"] = e.GitProviderConfig.ProviderId
		props["is_self_hosted"] = e.GitProviderConfig.BaseApiUrl != nil && *e.GitProviderConfig.BaseApiUrl != ""
		if e.GitProviderConfig.SigningMethod != nil {
			props["signing_method"] = *e.GitProviderConfig.SigningMethod
		}
	}

	return props
}
