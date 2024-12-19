// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var RunnerCmd = &cobra.Command{
	Use:     "runner",
	Short:   "Manage the runner",
	GroupID: util.RUNNER_GROUP,
}

func init() {
	RunnerCmd.AddCommand(configCmd)
	RunnerCmd.AddCommand(configureCmd)
	RunnerCmd.AddCommand(startCmd)
	RunnerCmd.AddCommand(serveCmd)
	RunnerCmd.AddCommand(daemonServeCmd)
	RunnerCmd.AddCommand(stopCmd)
	RunnerCmd.AddCommand(restartCmd)
	RunnerCmd.AddCommand(logsCmd)
	RunnerCmd.AddCommand(purgeCmd)
}
