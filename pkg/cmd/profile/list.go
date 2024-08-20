// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/profile"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profiles",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if format.FormatFlag != "" {
			formatter := format.NewFormatter(c.Profiles)
			formatter.Print()
			return
		}

		profile.ListProfiles(c.Profiles, c.ActiveProfileId)
	},
}

func init() {
	format.RegisterFormatFlag(profileListCmd)
}
