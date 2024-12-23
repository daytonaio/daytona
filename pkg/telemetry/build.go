// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "github.com/daytonaio/daytona/pkg/models"

type BuildEventName string

const (
	BuildEventLifecycleCreated             BuildEventName = "build_lifecycle_created"
	BuildEventLifecycleCreationFailed      BuildEventName = "build_lifecycle_creation_failed"
	BuildEventLifecycleDeleted             BuildEventName = "build_lifecycle_deleted"
	BuildEventLifecycleDeletionFailed      BuildEventName = "build_lifecycle_deletion_failed"
	BuildEventLifecycleForceDeleted        BuildEventName = "build_lifecycle_force_deleted"
	BuildEventLifecycleForceDeletionFailed BuildEventName = "build_lifecycle_force_deletion_failed"
)

type buildEvent struct {
	build *models.Build
	AbstractEvent
}

func NewBuildEvent(name BuildEventName, b *models.Build, err error, extras map[string]interface{}) Event {
	return buildEvent{
		build: b,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e buildEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.build != nil {
		props["build_id"] = e.build.Id
		props["from_prebuild"] = e.build.PrebuildId != ""
		if isImagePublic(e.build.ContainerConfig.Image) {
			props["image"] = e.build.Image
		}

		if e.build.Repository != nil && isPublic(e.build.Repository.Url) {
			props["repository_url"] = e.build.Repository.Url
		}

		props["builder"] = getBuilder(e.build.BuildConfig)
	}

	return props
}
