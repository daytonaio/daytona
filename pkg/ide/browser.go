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
	"github.com/daytonaio/daytona/pkg/build/devcontainer"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

const startVSCodeServerCommand = "$HOME/vscode-server/bin/openvscode-server --start-server --port=63000 --host=0.0.0.0 --without-connection-token --disable-workspace-trust --default-folder="

func OpenBrowserIDE(activeProfile config.Profile, targetId string, projectName string, projectProviderMetadata string, gpgKey string) error {
	// Download and start IDE
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, projectName, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Downloading OpenVSCode Server...")
	projectHostname := config.GetProjectHostname(activeProfile.Id, targetId, projectName)

	installServerCommand := exec.Command("ssh", projectHostname, "curl -fsSL https://download.daytona.io/daytona/get-openvscode-server.sh | sh")
	installServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
	installServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

	err = installServerCommand.Run()
	if err != nil {
		return err
	}

	projectDir, err := util.GetProjectDir(activeProfile, targetId, projectName, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Starting OpenVSCode Server...")

	go func() {
		startServerCommand := exec.CommandContext(context.Background(), "ssh", projectHostname, fmt.Sprintf("%s%s", startVSCodeServerCommand, projectDir))
		startServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
		startServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

		err = startServerCommand.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Forward IDE port
	browserPort, errChan := tailscale.ForwardPort(targetId, projectName, 63000, activeProfile)
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

	views.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s IDE port to %s.\nOpening browser...\n", projectName, ideURL))

	err = browser.OpenURL(ideURL)
	if err != nil {
		log.Error("Error opening URL: " + err.Error())
	}

	if projectProviderMetadata == "" {
		return nil
	}

	err = setupVSCodeCustomizations(projectHostname, projectProviderMetadata, devcontainer.Browser, "*/vscode-server/bin/openvscode-server", "$HOME/.openvscode-server/data/Machine/settings.json", ".daytona-customizations-lock-vscode-browser")
	if err != nil {
		log.Errorf("Error setting up IDE customizations: %s", err)
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
