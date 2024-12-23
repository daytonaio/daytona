// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/models"
)

type JobEventName string

const (
	JobEventLifecycleCreated        JobEventName = "job_lifecycle_created"
	JobEventLifecycleCreationFailed JobEventName = "job_lifecycle_creation_failed"
	JobEventRunStarted              JobEventName = "job_run_started"
	JobEventRunCompleted            JobEventName = "job_run_completed"
	JobEventRunFailed               JobEventName = "job_run_failed"
)

type jobEvent struct {
	AbstractEvent
	job *models.Job
}

func NewJobEvent(name JobEventName, j *models.Job, err error, extras map[string]interface{}) Event {
	return jobEvent{
		job: j,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e jobEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.job != nil {
		props["job_id"] = e.job.Id
		props["is_local_runner"] = e.job.RunnerId != nil && *e.job.RunnerId == common.LOCAL_RUNNER_ID
		props["resource_type"] = e.job.ResourceType
		props["job_action"] = e.job.Action

		if e.job.Error != nil {
			props["job_error"] = *e.job.Error
		}
	}

	return props
}
