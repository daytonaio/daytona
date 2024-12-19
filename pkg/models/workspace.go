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
	Name                string                     `json:"name" validate:"required" gorm:"not null"`
	Image               string                     `json:"image" validate:"required" gorm:"not null"`
	User                string                     `json:"user" validate:"required" gorm:"not null"`
	BuildConfig         *BuildConfig               `json:"buildConfig,omitempty" validate:"optional" gorm:"serializer:json"`
	Repository          *gitprovider.GitRepository `json:"repository" validate:"required" gorm:"serializer:json;not null"`
	EnvVars             map[string]string          `json:"envVars" validate:"required" gorm:"serializer:json;not null"`
	TargetId            string                     `json:"targetId" validate:"required" gorm:"not null"`
	Target              Target                     `json:"target" validate:"required" gorm:"foreignKey:TargetId"`
	ApiKey              string                     `json:"apiKey" validate:"required" gorm:"not null"`
	Metadata            *WorkspaceMetadata         `json:"metadata" validate:"optional" gorm:"foreignKey:WorkspaceId;references:Id"`
	GitProviderConfigId *string                    `json:"gitProviderConfigId,omitempty" validate:"optional"`
	LastJob             *Job                       `json:"lastJob" validate:"optional" gorm:"foreignKey:ResourceId;references:Id"`
	ProviderMetadata    *string                    `json:"providerMetadata,omitempty" validate:"optional"`
} // @name Workspace

type WorkspaceMetadata struct {
	WorkspaceId string     `json:"workspaceId" validate:"required" gorm:"primaryKey"`
	UpdatedAt   time.Time  `json:"updatedAt" validate:"required" gorm:"not null"`
	Uptime      uint64     `json:"uptime" validate:"required" gorm:"not null"`
	GitStatus   *GitStatus `json:"gitStatus" validate:"required" gorm:"serializer:json;not null"`
} // @name WorkspaceMetadata

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

type CachedBuild struct {
	User  string `json:"user" validate:"required"`
	Image string `json:"image" validate:"required"`
} // @name CachedBuild

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
