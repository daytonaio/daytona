// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type Workspace struct {
	Id                  string                     `json:"id" validate:"required" gorm:"primaryKey"`
	Name                string                     `json:"name" validate:"required"`
	Image               string                     `json:"image" validate:"required"`
	User                string                     `json:"user" validate:"required"`
	BuildConfig         *BuildConfig               `json:"buildConfig,omitempty" validate:"optional" gorm:"serializer:json"`
	Repository          *gitprovider.GitRepository `json:"repository" validate:"required" gorm:"serializer:json"`
	EnvVars             map[string]string          `json:"envVars" validate:"required" gorm:"serializer:json"`
	TargetId            string                     `json:"targetId" validate:"required" gorm:"foreignKey:TargetId;references:Id"`
	Target              Target                     `json:"target" validate:"required" gorm:"foreignKey:TargetId"`
	ApiKey              string                     `json:"-"`
	Metadata            *WorkspaceMetadata         `gorm:"foreignKey:WorkspaceId;references:Id" validate:"optional"`
	GitProviderConfigId *string                    `json:"gitProviderConfigId,omitempty" validate:"optional"`
	LastJob             *Job                       `gorm:"foreignKey:ResourceId;references:Id"`
} // @name Workspace

type WorkspaceMetadata struct {
	WorkspaceId string     `json:"workspaceId" validate:"required" gorm:"primaryKey;foreignKey:WorkspaceId;references:Id"`
	UpdatedAt   time.Time  `json:"updatedAt" validate:"required"`
	Uptime      uint64     `json:"uptime" validate:"required"`
	GitStatus   *GitStatus `json:"gitStatus" validate:"required" gorm:"serializer:json"`
} // @name WorkspaceMetadata

type ResourceState struct {
	Name      ResourceStateName `json:"name" validate:"required"`
	Error     *string           `json:"error" validate:"optional"`
	UpdatedAt time.Time         `json:"updatedAt" validate:"required"`
} // @name ResourceState

type ResourceStateName string

const (
	ResourceStateNameUndefined           ResourceStateName = "undefined"
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

func (w *Workspace) WorkspaceFolderName() string {
	if w.Repository != nil {
		return w.Repository.Name
	}
	return w.Name
}

func (w *Workspace) GetState() ResourceState {
	state := getResourceStateFromJob(w.LastJob)

	// If the workspace should be running, check if it is unresponsive
	if state.Name == ResourceStateNameStarted {
		if w.Metadata != nil && time.Since(w.Metadata.UpdatedAt) > AGENT_UNRESPONSIVE_THRESHOLD {
			state.Name = ResourceStateNameUnresponsive
			state.Error = util.Pointer("Workspace is unresponsive")
			state.UpdatedAt = w.Metadata.UpdatedAt
		}
	}

	return state
}

type BuildConfig struct {
	Devcontainer *DevcontainerConfig `json:"devcontainer,omitempty" validate:"optional"`
	CachedBuild  *CachedBuild        `json:"cachedBuild,omitempty" validate:"optional"`
} // @name BuildConfig

type DevcontainerConfig struct {
	FilePath string `json:"filePath" validate:"required"`
} // @name DevcontainerConfig

type CachedBuild struct {
	User  string `json:"user" validate:"required"`
	Image string `json:"image" validate:"required"`
} // @name CachedBuild

type WorkspaceInfo struct {
	Name             string `json:"name" validate:"required"`
	Created          string `json:"created" validate:"required"`
	IsRunning        bool   `json:"isRunning" validate:"required"`
	ProviderMetadata string `json:"providerMetadata,omitempty" validate:"optional"`
	TargetId         string `json:"targetId" validate:"required"`
} // @name WorkspaceInfo

type GitStatus struct {
	CurrentBranch   string        `json:"currentBranch" validate:"required"`
	Files           []*FileStatus `json:"fileStatus" validate:"required"`
	BranchPublished bool          `json:"branchPublished" validate:"optional"`
	Ahead           int           `json:"ahead" validate:"optional"`
	Behind          int           `json:"behind" validate:"optional"`
} // @name GitStatus

type FileStatus struct {
	Name     string `json:"name" validate:"required"`
	Extra    string `json:"extra" validate:"required"`
	Staging  Status `json:"staging" validate:"required"`
	Worktree Status `json:"worktree" validate:"required"`
} // @name FileStatus

// Status status code of a file in the Worktree
type Status string // @name Status

const (
	Unmodified         Status = "Unmodified"
	Untracked          Status = "Untracked"
	Modified           Status = "Modified"
	Added              Status = "Added"
	Deleted            Status = "Deleted"
	Renamed            Status = "Renamed"
	Copied             Status = "Copied"
	UpdatedButUnmerged Status = "Updated but unmerged"
)
