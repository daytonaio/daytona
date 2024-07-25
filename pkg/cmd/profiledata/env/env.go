// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var EnvCmd = &cobra.Command{
	Use:     "env",
	Short:   "Manage profile environment variables that are added to all workspaces",
	GroupID: util.PROFILE_GROUP,
}

func init() {
	EnvCmd.AddCommand(setCmd)
	EnvCmd.AddCommand(listCmd)
}
