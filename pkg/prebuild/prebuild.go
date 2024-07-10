// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import "github.com/daytonaio/daytona/pkg/workspace/project/config"

// Prebuild holds configuration for the prebuild process
type Prebuild struct {
	Key            string               `json:"key"`            // Composite key (project-config-name+branch-name) for the prebuild
	Branch         string               `json:"branch"`         // Branch to watch for changes
	ProjectConfig  config.ProjectConfig `json:"projectConfig"`  // Project configuration
	CommitInterval *int                 `json:"commitInterval"` // Number of commits between each new prebuild
	TriggerFiles   []string             `json:"triggerFiles"`   // Files that should trigger a new prebuild if changed
} // @name PrebuildConfig
