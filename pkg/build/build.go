// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import "github.com/daytonaio/daytona/pkg/workspace/project"

type BuildState string

const (
	BuildStatePending   BuildState = "pending"
	BuildStateRunning   BuildState = "running"
	BuildStateError     BuildState = "error"
	BuildStateSuccess   BuildState = "success"
	BuildStatePublished BuildState = "published"
)

type Build struct {
	Id      string          `json:"id" validate:"required"`
	Hash    string          `json:"hash" validate:"required"`
	State   BuildState      `json:"state" validate:"required"`
	Project project.Project `json:"project" validate:"required"`
	User    string          `json:"user" validate:"required"`
	Image   string          `json:"image" validate:"required"`
} // @name Build
