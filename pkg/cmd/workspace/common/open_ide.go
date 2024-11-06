// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

func OpenIDE(ideId string, activeProfile config.Profile, workspaceId string, workspaceProviderMetadata string, yesFlag bool, gpgKey string) error {
	telemetry.AdditionalData["ide"] = ideId

	switch ideId {
	case "vscode":
		return ide.OpenVSCode(activeProfile, workspaceId, workspaceProviderMetadata, gpgKey)
	case "ssh":
		return ide.OpenTerminalSsh(activeProfile, workspaceId, gpgKey, nil)
	case "browser":
		return ide.OpenBrowserIDE(activeProfile, workspaceId, workspaceProviderMetadata, gpgKey)
	case "cursor":
		return ide.OpenCursor(activeProfile, workspaceId, workspaceProviderMetadata, gpgKey)
	case "jupyter":
		return ide.OpenJupyterIDE(activeProfile, workspaceId, workspaceProviderMetadata, yesFlag, gpgKey)
	case "fleet":
		return ide.OpenFleet(activeProfile, workspaceId, gpgKey)
	case "zed":
		return ide.OpenZed(activeProfile, workspaceId, gpgKey)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			return ide.OpenJetbrainsIDE(activeProfile, ideId, workspaceId, gpgKey)
		}
	}

	return errors.New("invalid IDE. Please choose one by running `daytona ide`")
}
