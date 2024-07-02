// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"github.com/spf13/cobra"
)

var PrebuildCmd = &cobra.Command{
	Use:   "prebuild",
	Short: "Manage prebuilds",
}

func init() {
	PrebuildCmd.AddCommand(prebuildAddCmd)
}
