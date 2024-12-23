// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/spf13/cobra"
)

var RunnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Manage runners",
}

func init() {
	RunnerCmd.AddCommand(logsCmd)
	RunnerCmd.AddCommand(listCmd)
	RunnerCmd.AddCommand(registerCmd)
	RunnerCmd.AddCommand(unregisterCmd)
}
