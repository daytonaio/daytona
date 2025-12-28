// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package gitprovider

type CloneTarget string // @name CloneTarget

const (
	CloneTargetBranch CloneTarget = "branch"
	CloneTargetCommit CloneTarget = "commit"
)

type GitRepository struct {
	Id       string      `json:"id" validate:"required"`
	Url      string      `json:"url" validate:"required"`
	Name     string      `json:"name" validate:"required"`
	Branch   string      `json:"branch" validate:"required"`
	Sha      string      `json:"sha" validate:"required"`
	Owner    string      `json:"owner" validate:"required"`
	PrNumber *uint32     `json:"prNumber,omitempty" validate:"optional"`
	Source   string      `json:"source" validate:"required"`
	Path     *string     `json:"path,omitempty" validate:"optional"`
	Target   CloneTarget `json:"cloneTarget,omitempty" validate:"optional"`
} // @name GitRepository
