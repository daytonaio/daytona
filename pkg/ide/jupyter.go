// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"context"
	"fmt"
	"io"
	"os"
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

const startJupyterCommand = "notebook --no-browser --port=8888 --ip=0.0.0.0 --NotebookApp.token='' --NotebookApp.password=''"

func OpenJupyterIDE(activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string) error {
	// Download and start IDE
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Installing Jupyter Notebook...")
	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	// Install Jupyter Notebook
	installJupyterCommand := exec.Command("ssh", projectHostname, "python3 -m pip install --user notebook && export PATH=$HOME/.local/bin:$PATH && echo $PATH")
	installJupyterCommand.Stdout = io.Writer(&util.DebugLogWriter{})
	installJupyterCommand.Stderr = io.Writer(&util.DebugLogWriter{})

	err = installJupyterCommand.Run()
	if err != nil {
		return fmt.Errorf("failed to install Jupyter Notebook: %w", err)
	}

	// Check Jupyter Notebook installation and print PATH
	checkCommand := exec.Command("ssh", projectHostname, "export PATH=$HOME/.local/bin:$PATH && echo $PATH && $HOME/.local/bin/jupyter notebook --version")
	checkCommand.Stdout = os.Stdout
	checkCommand.Stderr = os.Stderr
	if err := checkCommand.Run(); err != nil {
		return fmt.Errorf("failed to check Jupyter Notebook installation: %w", err)
	}

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Starting Jupyter Notebook...")

	go func() {
		startServerCommand := exec.CommandContext(context.Background(), "ssh", projectHostname, fmt.Sprintf("export PATH=$HOME/.local/bin:$PATH && cd %s && jupyter %s", projectDir, startJupyterCommand))
		startServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
		startServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

		err = startServerCommand.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Forward IDE port
	browserPort, errChan := tailscale.ForwardPort(workspaceId, projectName, 8888, activeProfile)
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

	views.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s Jupyter Notebook port to %s.\nOpening browser...\n", projectName, ideURL))

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
