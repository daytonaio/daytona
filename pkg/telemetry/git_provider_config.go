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
	gitProviderConfig *models.GitProviderConfig
	AbstractEvent
}

func NewGitProviderConfigEvent(name GitProviderConfigEventName, b *models.GitProviderConfig, err error, extras map[string]interface{}) Event {
	return GitProviderConfigEvent{
		gitProviderConfig: b,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e GitProviderConfigEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.gitProviderConfig == nil {
		return props
	}

	props["provider_id"] = e.gitProviderConfig.ProviderId
	props["is_self_hosted"] = e.gitProviderConfig.BaseApiUrl != nil && *e.gitProviderConfig.BaseApiUrl != ""
	if e.gitProviderConfig.SigningMethod != nil {
		props["signing_method"] = *e.gitProviderConfig.SigningMethod
	}

	return props
}
