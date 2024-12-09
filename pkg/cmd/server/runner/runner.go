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
	RunnerCmd.AddCommand(runnerListCmd)
	RunnerCmd.AddCommand(runnerRegisterCmd)
	RunnerCmd.AddCommand(runnerUnregisterCmd)
}
