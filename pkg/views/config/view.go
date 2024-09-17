// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strconv"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/profile"
)

func Render(cfg *config.Config, showApiKeysFlag bool) {
	output := "\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("ID: "), cfg.Id) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Default IDE: "), cfg.DefaultIdeId) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Telemetry Enabled: "), strconv.FormatBool(cfg.TelemetryEnabled)) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Active Profile: "), cfg.ActiveProfileId) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Profiles: "), len(cfg.Profiles)) + "\n\n"

	profiles, err := profile.ListProfiles(cfg.Profiles, cfg.ActiveProfileId, showApiKeysFlag)
	if err != nil {
		fmt.Print(output)
		return
	}

	output += profiles

	fmt.Print(output)
}
