// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var ProfileCmd = &cobra.Command{
	Use:     "profile",
	Short:   "Manage profiles",
	GroupID: util.PROFILE_GROUP,
}

func init() {
	ProfileCmd.AddGroup(&cobra.Group{ID: util.PROFILE_GROUP, Title: "Profile"})
	ProfileCmd.AddCommand(profileListCmd)
	ProfileCmd.AddCommand(ProfileUseCmd)
	ProfileCmd.AddCommand(ProfileAddCmd)
	ProfileCmd.AddCommand(profileEditCmd)
	ProfileCmd.AddCommand(profileDeleteCmd)
}
