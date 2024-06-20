// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import "fmt"

type BuilderChoice string

const (
	AUTOMATIC    BuilderChoice = "auto"
	DEVCONTAINER BuilderChoice = "devcontainer"
	CUSTOMIMAGE  BuilderChoice = "custom-image"
	NONE         BuilderChoice = "none"
)

// String is used both by fmt.Print and by Cobra in help text
func (c *BuilderChoice) String() string {
	return string(*c)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (c *BuilderChoice) Set(v string) error {
	switch v {
	case string(AUTOMATIC), string(DEVCONTAINER), string(CUSTOMIMAGE), string(NONE):
		*c = BuilderChoice(v)
		return nil
	default:
		return fmt.Errorf("Build type must be one of %s/%s/%s", AUTOMATIC, DEVCONTAINER, NONE)
	}
}

// Type is only used in help text
func (c *BuilderChoice) Type() string {
	return "BuildChoice"
}
