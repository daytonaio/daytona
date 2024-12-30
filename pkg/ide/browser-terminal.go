// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

const startCommand = "$HOME/ttyd/bin/ttyd --port 63777 --writable --cwd"

func OpenBrowserTerminal(activeProfile config.Profile, workspaceId string, projectName string, gpgKey string) error {
	// Make sure SSH config exists
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Downloading Terminal Server...")
	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	// Download and start ttyd
	installServerCommand := exec.Command("ssh", projectHostname, "curl -fsSL https://raw.githubusercontent.com/hunnywar/daytona/f9ec6d34bf5c8b4b405cb4979483b066d922cf55/hack/get-ttyd.sh | sh")
	installServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
	installServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

	err = installServerCommand.Run()
	if err != nil {
		return err
	}

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Starting Terminal Server...")

	go func() {
		startServerCommand := exec.CommandContext(context.Background(), "ssh", projectHostname, fmt.Sprintf("%s %s bash", startCommand, projectDir))
		startServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
		startServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

		err = startServerCommand.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Forward ttyd (Terminal server) port
	browserPort, errChan := tailscale.ForwardPort(workspaceId, projectName, 63777, activeProfile)
	if browserPort == nil {
		if err := <-errChan; err != nil {
			return err
		}
	}

	ideURL := fmt.Sprintf("http://localhost:%d", *browserPort)
	// Wait for the port to be ready
	for {
		if ports.IsPortReady(*browserPort) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	views.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s Terminal port to %s.\nOpening browser...\n", projectName, ideURL))

	err = browser.OpenURL(ideURL)
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
