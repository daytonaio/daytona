// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import "time"

type ResourceState struct {
	Name      ResourceStateName `json:"name" validate:"required" gorm:"not null"`
	Error     *string           `json:"error" validate:"optional"`
	UpdatedAt time.Time         `json:"updatedAt" validate:"required" gorm:"not null"`
} // @name ResourceState

type ResourceStateName string

const (
	ResourceStateNameUndefined           ResourceStateName = "undefined"
	ResourceStateNamePendingRun          ResourceStateName = "pending-run"
	ResourceStateNameRunning             ResourceStateName = "running"
	ResourceStateNameRunSuccessful       ResourceStateName = "run-successful"
	ResourceStateNamePendingCreate       ResourceStateName = "pending-create"
	ResourceStateNameCreating            ResourceStateName = "creating"
	ResourceStateNamePendingStart        ResourceStateName = "pending-start"
	ResourceStateNameStarting            ResourceStateName = "starting"
	ResourceStateNameStarted             ResourceStateName = "started"
	ResourceStateNamePendingStop         ResourceStateName = "pending-stop"
	ResourceStateNameStopping            ResourceStateName = "stopping"
	ResourceStateNameStopped             ResourceStateName = "stopped"
	ResourceStateNamePendingRestart      ResourceStateName = "pending-restart"
	ResourceStateNameError               ResourceStateName = "error"
	ResourceStateNameUnresponsive        ResourceStateName = "unresponsive"
	ResourceStateNamePendingDelete       ResourceStateName = "pending-delete"
	ResourceStateNamePendingForcedDelete ResourceStateName = "pending-forced-delete"
	ResourceStateNameDeleting            ResourceStateName = "deleting"
	ResourceStateNameDeleted             ResourceStateName = "deleted"
)

type BuildConfig struct {
	Devcontainer *DevcontainerConfig `json:"devcontainer,omitempty" validate:"optional"`
	CachedBuild  *CachedBuild        `json:"cachedBuild,omitempty" validate:"optional"`
} // @name BuildConfig

type DevcontainerConfig struct {
	FilePath string `json:"filePath" validate:"required" gorm:"not null"`
} // @name DevcontainerConfig
