// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceEventName string

const (
	WorkspaceEventLifecycleCreated             WorkspaceEventName = "workspace_lifecycle_created"
	WorkspaceEventLifecycleCreationFailed      WorkspaceEventName = "workspace_lifecycle_creation_failed"
	WorkspaceEventLifecycleStarted             WorkspaceEventName = "workspace_lifecycle_started"
	WorkspaceEventLifecycleStartFailed         WorkspaceEventName = "workspace_lifecycle_start_failed"
	WorkspaceEventLifecycleRestarted           WorkspaceEventName = "workspace_lifecycle_restarted"
	WorkspaceEventLifecycleRestartFailed       WorkspaceEventName = "workspace_lifecycle_restart_failed"
	WorkspaceEventLifecycleStopped             WorkspaceEventName = "workspace_lifecycle_stopped"
	WorkspaceEventLifecycleStopFailed          WorkspaceEventName = "workspace_lifecycle_stop_failed"
	WorkspaceEventLifecycleDeleted             WorkspaceEventName = "workspace_lifecycle_deleted"
	WorkspaceEventLifecycleDeletionFailed      WorkspaceEventName = "workspace_lifecycle_deletion_failed"
	WorkspaceEventLifecycleForceDeleted        WorkspaceEventName = "workspace_lifecycle_force_deleted"
	WorkspaceEventLifecycleForceDeletionFailed WorkspaceEventName = "workspace_lifecycle_force_deletion_failed"
	WorkspaceEventLabelsUpdated                WorkspaceEventName = "workspace_labels_updated"
	WorkspaceEventLabelsUpdateFailed           WorkspaceEventName = "workspace_label_update_failed"
)

type workspaceEvent struct {
	AbstractEvent
	workspace *models.Workspace
}

func NewWorkspaceEvent(name WorkspaceEventName, w *models.Workspace, err error, extras map[string]interface{}) Event {
	return workspaceEvent{
		workspace: w,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e workspaceEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.workspace == nil {
		return props
	}

	props["workspace_id"] = e.workspace.Id
	props["is_local_docker_target_config"] = common.IsLocalDockerTarget(e.workspace.Target.TargetConfig.ProviderInfo.Name, e.workspace.Target.TargetConfig.Options, e.workspace.Target.TargetConfig.ProviderInfo.RunnerId)
	props["provider_name"] = e.workspace.Target.TargetConfig.ProviderInfo.Name
	props["provider_version"] = e.workspace.Target.TargetConfig.ProviderInfo.Version
	if isImagePublic(e.workspace.Image) {
		props["image"] = e.workspace.Image
	}
	if e.workspace.Repository != nil && isPublic(e.workspace.Repository.Url) {
		props["repository_url"] = e.workspace.Repository.Url
	}
	props["builder"] = getBuilder(e.workspace.BuildConfig)
	props["is_local_runner"] = e.workspace.Target.TargetConfig.ProviderInfo.RunnerId == common.LOCAL_RUNNER_ID
	props["n_labels"] = len(e.workspace.Labels)

	return props
}

func getBuilder(bc *models.BuildConfig) string {
	if bc == nil {
		return "none"
	} else if bc.Devcontainer != nil {
		return "devcontainer"
	} else {
		return "automatic"
	}
}
