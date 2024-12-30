// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"

	"github.com/daytonaio/daytona/pkg/cmd/format"
	view "github.com/daytonaio/daytona/pkg/views/whoami"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "Display information about the active user",
	Args:    cobra.NoArgs,
	Aliases: []string{"who", "user"},
	GroupID: util.PROFILE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		profile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(profile)
			formattedData.Print()
			return nil
		}

		view.Render(profile)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(whoamiCmd)
}
