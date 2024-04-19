// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package whoami

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
)

const listLabelWidth = 20

func Render(profile config.Profile) {
	var output string

	if profile.Id == "default" {
		output += views.GetBoldedInfoMessage("You are currently on the default profile") + "\n"
	} else {
		output += views.GetBoldedInfoMessage("You are currently on profile "+profile.Name) + "\n"
	}
	output += views.GetListLine(fmt.Sprintf("\x1b[1m%-*s\x1b[0m%s", listLabelWidth, "Profile ID:", profile.Id)) + "\n"

	if profile.Api.Url != "" {
		output += views.GetListLine(fmt.Sprintf("\x1b[1m%-*s\x1b[0m%s", listLabelWidth, "API URL:", profile.Api.Url))
	}

	views.RenderContainerLayout(output)
}
