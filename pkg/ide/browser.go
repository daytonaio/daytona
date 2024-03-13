// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/ports"
	view_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

const startVSCodeServerCommand = "$HOME/vscode-server/bin/openvscode-server --start-server --port=63000 --host=0.0.0.0 --without-connection-token --disable-workspace-trust --default-folder=$DAYTONA_WS_DIR"

func OpenBrowserIDE(activeProfile config.Profile, workspaceId string, projectName string) error {
	// Download and start IDE
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return err
	}

	view_util.RenderInfoMessageBold("Downloading OpenVSCode Server...")
	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	installServerCommand := exec.Command("ssh", projectHostname, "curl -fsSL https://download.daytona.io/daytona/get-openvscode-server.sh | sh")
	installServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
	installServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

	err = installServerCommand.Run()
	if err != nil {
		return err
	}

	view_util.RenderInfoMessageBold("Starting OpenVSCode Server...")

	go func() {
		startServerCommand := exec.CommandContext(context.Background(), "ssh", projectHostname, startVSCodeServerCommand)
		startServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
		startServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

		err = startServerCommand.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Forward IDE port
	browserPort, errChan := ports.ForwardPort(workspaceId, projectName, 63000)
	if browserPort == nil {
		if err := <-errChan; err != nil {
			return err
		}
	}

	view_util.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s IDE port to %d.\nOpening browser...", projectName, *browserPort))

	err = browser.OpenURL(fmt.Sprintf("http://localhost:%d", *browserPort))
	if err != nil {
		log.Error("Error opening URL: " + err.Error())
	}

	for {
		err := <-errChan
		if err != nil {
			// Log only in debug mode
			// Connection errors to the forwarded port should not exit the process
			log.Debug(err)
		}
	}
}
