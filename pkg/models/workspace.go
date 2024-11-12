// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"strings"

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
	State               *WorkspaceState            `json:"state,omitempty" validate:"optional" gorm:"serializer:json"`
	GitProviderConfigId *string                    `json:"gitProviderConfigId,omitempty" validate:"optional"`
} // @name Workspace

func (w *Workspace) WorkspaceFolderName() string {
	if w.Repository != nil {
		return w.Repository.Name
	}
	return w.Name
}

func (w Workspace) Hostname() string {
	// Replace special chars with hyphen to form valid hostname
	// String resulting in consecutive hyphens is also valid
	workspaceId := w.Id
	workspaceId = strings.ReplaceAll(workspaceId, "_", "-")
	workspaceId = strings.ReplaceAll(workspaceId, "*", "-")
	workspaceId = strings.ReplaceAll(workspaceId, ".", "-")

	hostname := fmt.Sprintf("ws-%s", workspaceId)

	if len(hostname) > 63 {
		return hostname[:63]
	}

	return hostname
}

type WorkspaceInfo struct {
	Name             string `json:"name" validate:"required"`
	Created          string `json:"created" validate:"required"`
	IsRunning        bool   `json:"isRunning" validate:"required"`
	ProviderMetadata string `json:"providerMetadata,omitempty" validate:"optional"`
	TargetId         string `json:"targetId" validate:"required"`
} // @name WorkspaceInfo

type WorkspaceState struct {
	UpdatedAt string     `json:"updatedAt" validate:"required"`
	Uptime    uint64     `json:"uptime" validate:"required"`
	GitStatus *GitStatus `json:"gitStatus" validate:"required"`
} // @name WorkspaceState

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
