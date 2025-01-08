// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/exec"
	"sync"
	"syscall"
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

// OpenBrowserTerminal starts a browser-based terminal and opens it in the browser
func OpenBrowserTerminal(activeProfile config.Profile, workspaceId string, gpgKey *string) error {
	// Create a cancellation context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the context is canceled on exit

	// Capture OS interrupt signals for graceful exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Ensure SSH config exists
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Downloading Terminal Server...")
	workspaceHostname := config.GetHostname(activeProfile.Id, workspaceId)

	// Download and install ttyd
	installServerCommand := exec.Command("ssh", workspaceHostname, "curl -fsSL https://download.daytona.io/daytona/tools/get-ttyd.sh | sh")
	installServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
	installServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

	err = installServerCommand.Run()
	if err != nil {
		return err
	}

	workspaceDir, err := util.GetWorkspaceDir(activeProfile, workspaceId, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Starting Terminal Server...")

	// Start the terminal server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		startServerCommand := exec.CommandContext(ctx, "ssh", workspaceHostname, fmt.Sprintf("%s %s bash", startCommand, workspaceDir))
		startServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
		startServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

		err = startServerCommand.Run()
		if err != nil && ctx.Err() == nil { // Ignore errors if context was canceled
			log.Error(err)
		}
	}()

	// Forward ttyd (Terminal server) port
	browserPort, errChan := tailscale.ForwardPort(workspaceId, 63777, activeProfile)
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
		time.Sleep(100 * time.Millisecond)
	}

	views.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s Terminal port to %s.\nOpening browser...\n", workspaceId, ideURL))

	err = browser.OpenURL(ideURL)
	if err != nil {
		log.Error("Error opening URL: " + err.Error())
	}

	// Handle errors from the port-forwarding goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errChan:
				if err != nil {
					// Log errors in debug mode
					// Connection errors to the forwarded port should not exit the process
					log.Debug(err)
				}
			}
		}
	}()

	// Wait for a termination signal
	<-signalChan
	log.Info("Received termination signal. Shutting down gracefully...")

	// Cancel the context to stop all goroutines
	cancel()

	// Wait for all goroutines to complete
	wg.Wait()
	log.Info("All tasks stopped. Exiting.")
	return nil
}