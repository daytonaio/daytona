// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/profile"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all profiles",
	Args:    cobra.NoArgs,
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if common.FormatFlag != "" {
			formattedData := common.NewFormatter(c.Profiles)
			formattedData.Print()
			return nil
		}

		activeProfileId := c.ActiveProfileId
		profile.ListProfiles(c.Profiles, &activeProfileId)
		return nil
	},
}

func init() {
	common.RegisterFormatFlag(ListCmd)
}
