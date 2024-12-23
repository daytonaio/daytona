// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/models"
)

type TargetEventName string

const (
	TargetEventLifecycleCreated             TargetEventName = "target_lifecycle_created"
	TargetEventLifecycleCreationFailed      TargetEventName = "target_lifecycle_creation_failed"
	TargetEventLifecycleStarted             TargetEventName = "target_lifecycle_started"
	TargetEventLifecycleStartFailed         TargetEventName = "target_lifecycle_start_failed"
	TargetEventLifecycleRestarted           TargetEventName = "target_lifecycle_restarted"
	TargetEventLifecycleRestartFailed       TargetEventName = "target_lifecycle_restart_failed"
	TargetEventLifecycleStopped             TargetEventName = "target_lifecycle_stopped"
	TargetEventLifecycleStopFailed          TargetEventName = "target_lifecycle_stop_failed"
	TargetEventLifecycleDeleted             TargetEventName = "target_lifecycle_deleted"
	TargetEventLifecycleDeletionFailed      TargetEventName = "target_lifecycle_deletion_failed"
	TargetEventLifecycleForceDeleted        TargetEventName = "target_lifecycle_force_deleted"
	TargetEventLifecycleForceDeletionFailed TargetEventName = "target_lifecycle_force_deletion_failed"
)

type targetEvent struct {
	AbstractEvent
	target *models.Target
}

func NewTargetEvent(name TargetEventName, t *models.Target, err error, extras map[string]interface{}) Event {
	return targetEvent{
		target: t,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e targetEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.target != nil {
		props["target_id"] = e.target.Id
		props["is_local_docker_target_config"] = common.IsLocalDockerTarget(e.target.TargetConfig.ProviderInfo.Name, e.target.TargetConfig.Options, e.target.TargetConfig.ProviderInfo.RunnerId)
		props["provider_name"] = e.target.TargetConfig.ProviderInfo.Name
		props["provider_version"] = e.target.TargetConfig.ProviderInfo.Version
		props["target_config_deleted"] = e.target.TargetConfig.Deleted
		props["agentless_target"] = e.target.TargetConfig.ProviderInfo.AgentlessTarget
		props["is_local_runner"] = e.target.TargetConfig.ProviderInfo.RunnerId == common.LOCAL_RUNNER_ID
	}

	return props
}
