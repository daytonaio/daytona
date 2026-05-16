// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cli/config"
	viewsprofile "github.com/daytonaio/daytona/cli/views/profile"
	"github.com/spf13/cobra"
)

// rootCommand if invoked without arguments takes user into the interactive
// profile management flow.
func rootCommand(_ *cobra.Command, _ []string) error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	if len(c.Profiles) == 0 {
		fmt.Println("No profiles found. Run `daytona login` to create one.")
		return nil
	}

	return viewsprofile.SelectProfile(
		c,
		c.Profiles,
		c.ActiveProfileId,
		setActiveProfile,
		editProfile,
		deleteProfile,
	)
}

func setActiveProfile(c *config.Config, p config.Profile) error {
	c.ActiveProfileId = p.Id
	return c.Save()
}

func editProfile(c *config.Config, p config.Profile) error {
	return c.EditProfile(p)
}

func deleteProfile(c *config.Config, p config.Profile) error {
	return c.RemoveProfile(p.Id)
}
