// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func OpenIDE(ideId string, activeProfile config.Profile, workspaceId, repoName, workspaceProviderMetadata string, yesFlag bool, gpgKey *string) error {
	var err error
	switch ideId {
	case "vscode":
		err = ide.OpenVSCode(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "code-insiders":
		err = ide.OpenVSCodeInsiders(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "ssh":
		err = ide.OpenTerminalSsh(activeProfile, workspaceId, gpgKey, nil)
	case "browser":
		err = ide.OpenBrowserIDE(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "codium":
		err = ide.OpenVScodium(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "codium-insiders":
		err = ide.OpenVScodiumInsiders(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "cursor":
		err = ide.OpenCursor(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "jupyter":
		err = ide.OpenJupyterIDE(activeProfile, workspaceId, repoName, workspaceProviderMetadata, yesFlag, gpgKey)
	case "fleet":
		err = ide.OpenFleet(activeProfile, workspaceId, repoName, gpgKey)
	case "positron":
		err = ide.OpenPositron(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	case "zed":
		err = ide.OpenZed(activeProfile, workspaceId, repoName, gpgKey)
	case "windsurf":
		err = ide.OpenWindsurf(activeProfile, workspaceId, repoName, workspaceProviderMetadata, gpgKey)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			err = ide.OpenJetbrainsIDE(activeProfile, ideId, workspaceId, repoName, gpgKey)
		} else {
			return errors.New("invalid IDE. Please choose one by running `daytona ide`")
		}
	}

	eventName := telemetry.CliEventWorkspaceOpened
	if err != nil {
		eventName = telemetry.CliEventWorkspaceOpenFailed
	}

	event := telemetry.NewCliEvent(eventName, nil, []string{}, err, map[string]interface{}{"ide": ideId})
	telemetryErr := TrackTelemetryEvent(event, config.GetClientId())
	if telemetryErr != nil {
		log.Trace(telemetryErr)
	}

	return err
}
