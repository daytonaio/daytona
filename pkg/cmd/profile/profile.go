// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/spf13/cobra"
)

var ProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
	ProfileCmd.AddCommand(ProfileUseCmd)
	ProfileCmd.AddCommand(ProfileAddCmd)
	ProfileCmd.AddCommand(profileEditCmd)
	ProfileCmd.AddCommand(profileDeleteCmd)
}
