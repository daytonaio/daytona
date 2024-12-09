// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/spf13/cobra"
)

var stopRunnerCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the runner",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
