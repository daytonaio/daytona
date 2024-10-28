// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/apiclient"
)

type BuildChoice string

const (
	AUTOMATIC    BuildChoice = "auto"
	DEVCONTAINER BuildChoice = "devcontainer"
	CUSTOMIMAGE  BuildChoice = "custom-image"
	NONE         BuildChoice = "none"
)

type WorkspaceConfigDefaults struct {
	BuildChoice          BuildChoice
	Image                *string
	ImageUser            *string
	DevcontainerFilePath string
}

func GetWorkspaceBuildChoice(workspace apiclient.CreateWorkspaceDTO, defaults *WorkspaceConfigDefaults) (BuildChoice, string) {
	if workspace.BuildConfig == nil {
		if workspace.Image != nil && workspace.User != nil &&
			defaults.Image != nil && defaults.ImageUser != nil &&
			*workspace.Image == *defaults.Image &&
			*workspace.User == *defaults.ImageUser {
			return NONE, "None"
		} else {
			return CUSTOMIMAGE, "Custom Image"
		}
	} else {
		if workspace.BuildConfig.Devcontainer != nil {
			return DEVCONTAINER, "Devcontainer"
		} else {
			return AUTOMATIC, "Automatic"
		}
	}
}

// String is used both by fmt.Print and by Cobra in help text
func (c *BuildChoice) String() string {
	return string(*c)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (c *BuildChoice) Set(v string) error {
	switch v {
	case string(AUTOMATIC), string(DEVCONTAINER), string(CUSTOMIMAGE), string(NONE):
		*c = BuildChoice(v)
		return nil
	default:
		return fmt.Errorf("Build type must be one of %s/%s/%s", AUTOMATIC, DEVCONTAINER, NONE)
	}
}

// Type is only used in help text
func (c *BuildChoice) Type() string {
	return "BuildChoice"
}
