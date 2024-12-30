// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/profile"

	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profiles",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if len(c.Profiles) == 0 {
			return config.ErrNoProfilesFound
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(c.Profiles)
			formattedData.Print()
			return nil
		}

		output, err := profile.ListProfiles(c.Profiles, c.ActiveProfileId, false)
		if err != nil {
			return err
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(profileListCmd)
}
