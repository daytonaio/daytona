// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
)

func GetProfileFromPrompt(profileList []config.Profile) (*config.Profile, error) {
	var chosenProfileId string
	var profileOptions []huh.Option[string]

	for _, profile := range profileList {
		profileOptions = append(profileOptions, huh.NewOption(profile.Name, profile.Id))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Profile").
				Options(
					profileOptions...,
				).
				Value(&chosenProfileId),
		).WithTheme(common.GetCustomTheme()),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	for _, profile := range profileList {
		if profile.Id == chosenProfileId {
			return &profile, nil
		}
	}

	return nil, nil
}
