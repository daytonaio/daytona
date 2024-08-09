// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"builds"},
	Short:   "Manage builds",
}

func init() {
	BuildCmd.AddCommand(buildListCmd)
}
