// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"
)

type Job struct {
	Id           string       `json:"id" validate:"required" gorm:"primaryKey"`
	ResourceId   string       `json:"resourceId" validate:"required"`
	ResourceType ResourceType `json:"resourceType" validate:"required"`
	State        JobState     `json:"state" validate:"required"`
	Action       JobAction    `json:"action" validate:"required"`
	Error        *string      `json:"error" validate:"optional"`
	CreatedAt    time.Time    `json:"createdAt" validate:"required"`
	UpdatedAt    time.Time    `json:"updatedAt" validate:"required"`
} // @name Job

type ResourceType string // @name ResourceType

const (
	ResourceTypeWorkspace ResourceType = "workspace"
	ResourceTypeTarget    ResourceType = "target"
	ResourceTypeBuild     ResourceType = "build"
)

type JobState string // @name JobState

const (
	JobStatePending JobState = "pending"
	JobStateRunning JobState = "running"
	JobStateError   JobState = "error"
	JobStateSuccess JobState = "success"
)

type JobAction string

const (
	JobActionCreate      JobAction = "create"
	JobActionStart       JobAction = "start"
	JobActionStop        JobAction = "stop"
	JobActionRestart     JobAction = "restart"
	JobActionDelete      JobAction = "delete"
	JobActionForceDelete JobAction = "force-delete"
	JobActionRun         JobAction = "run"
)

func getResourceStateFromJob(job *Job) ResourceState {
	state := ResourceState{
		Name:      ResourceStateNameUnresponsive,
		UpdatedAt: time.Now(),
	}

	if job == nil {
		return state
	}

	if job.State == JobStateSuccess {
		switch job.Action {
		case JobActionRun:
			state.Name = ResourceStateNameRunSuccessful
		case JobActionCreate:
			state.Name = ResourceStateNameStarted
		case JobActionStart:
			state.Name = ResourceStateNameStarted
		case JobActionStop:
			state.Name = ResourceStateNameStopped
		case JobActionRestart:
			state.Name = ResourceStateNameStarted
		case JobActionDelete:
			state.Name = ResourceStateNameDeleted
		case JobActionForceDelete:
			state.Name = ResourceStateNameDeleted
		}
	} else if job.State == JobStateError {
		state.Name = ResourceStateNameError
		state.Error = job.Error
	} else if job.State == JobStateRunning {
		switch job.Action {
		case JobActionRun:
			state.Name = ResourceStateNameRunning
		case JobActionCreate:
			state.Name = ResourceStateNameCreating
		case JobActionStart:
			state.Name = ResourceStateNameStarting
		case JobActionStop:
			state.Name = ResourceStateNameStopping
		case JobActionRestart:
			state.Name = ResourceStateNameStarting
		case JobActionDelete:
			state.Name = ResourceStateNameDeleting
		case JobActionForceDelete:
			state.Name = ResourceStateNameDeleting
		}
	}

	state.UpdatedAt = job.UpdatedAt
	return state
}
