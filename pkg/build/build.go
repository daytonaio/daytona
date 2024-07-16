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
	Id      string          `json:"id"`
	Hash    string          `json:"hash"`
	State   BuildState      `json:"state"`
	Project project.Project `json:"project"`
	User    string          `json:"user"`
	Image   string          `json:"image"`
} // @name Build
