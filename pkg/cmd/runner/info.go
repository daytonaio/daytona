// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/spf13/cobra"
)

var infoRunnerCmd = &cobra.Command{
	Use:   "info",
	Short: "View runner info",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
