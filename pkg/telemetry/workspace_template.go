// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceTemplateEventName string

var (
	WorkspaceTemplateEventLifecycleSaved          WorkspaceTemplateEventName = "workspace_template_lifecycle_saved"
	WorkspaceTemplateEventLifecycleSaveFailed     WorkspaceTemplateEventName = "workspace_template_lifecycle_save_failed"
	WorkspaceTemplateEventLifecycleDeleted        WorkspaceTemplateEventName = "workspace_template_lifecycle_deleted"
	WorkspaceTemplateEventLifecycleDeletionFailed WorkspaceTemplateEventName = "workspace_template_lifecycle_deletion_failed"
	WorkspaceTemplateEventPrebuildSaved           WorkspaceTemplateEventName = "workspace_template_prebuild_saved"
	WorkspaceTemplateEventPrebuildSaveFailed      WorkspaceTemplateEventName = "workspace_template_prebuild_save_failed"
	WorkspaceTemplateEventPrebuildDeleted         WorkspaceTemplateEventName = "workspace_template_prebuild_deleted"
	WorkspaceTemplateEventPrebuildDeletionFailed  WorkspaceTemplateEventName = "workspace_template_prebuild_deletion_failed"
)

type workspaceTemplateEvent struct {
	AbstractEvent
	workspaceTemplate *models.WorkspaceTemplate
}

func NewWorkspaceTemplateEvent(name WorkspaceTemplateEventName, wt *models.WorkspaceTemplate, err error, extras map[string]interface{}) Event {
	return workspaceTemplateEvent{
		workspaceTemplate: wt,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e workspaceTemplateEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.workspaceTemplate != nil {
		props["workspace_template_name"] = e.workspaceTemplate.Name
		// prebuilds, err := json.Marshal(e.workspaceTemplate.Prebuilds)
		props["prebuilds"] = e.workspaceTemplate.Prebuilds
		// if err == nil {
		// }
		if isImagePublic(e.workspaceTemplate.Image) {
			props["image"] = e.workspaceTemplate.Image
		}

		if isPublic(e.workspaceTemplate.RepositoryUrl) {
			props["repository_url"] = e.workspaceTemplate.RepositoryUrl
		}

		props["builder"] = getBuilder(e.workspaceTemplate.BuildConfig)
	}

	return props
}
