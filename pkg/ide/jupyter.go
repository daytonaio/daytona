// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"context"
	"errors"
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

	"github.com/charmbracelet/huh"
	log "github.com/sirupsen/logrus"
)

const startJupyterCommand = "notebook --no-browser --port=8888 --ip=0.0.0.0 --NotebookApp.token='' --NotebookApp.password=''"

// OpenJupyterIDE manages the installation and startup of a Jupyter IDE on a remote target.
func OpenJupyterIDE(activeProfile config.Profile, targetId, projectName, projectProviderMetadata string, yesFlag bool, gpgKey string) error {
	// Ensure SSH config entry is added
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, projectName, gpgKey)
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, targetId, projectName)

	// Check and install Python if necessary
	if err := ensurePythonInstalled(projectHostname, yesFlag); err != nil {
		return err
	}

	// Check and install pip if necessary
	if err := ensurePipInstalled(projectHostname, yesFlag); err != nil {
		return err
	}

	// Check and install Jupyter Notebook if necessary
	if err := ensureJupyterInstalled(projectHostname); err != nil {
		return err
	}

	// Start Jupyter Notebook server
	if err := startJupyterServer(projectHostname, activeProfile, targetId, projectName, gpgKey); err != nil {
		return err
	}

	return nil
}

// ensurePythonInstalled checks if Python is installed and installs it if the user agrees.
func ensurePythonInstalled(hostname string, yesFlag bool) error {
	views.RenderInfoMessageBold("Checking Python installation...")

	// Check if Python is installed
	if err := runRemoteCommand(hostname, "python3 --version"); err != nil {
		if yesFlag {
			return installPython(hostname)
		}

		var confirmInstall bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Would you like to install Python3?").
					Description("Python3 is required to run Jupyter Notebook.").
					Value(&confirmInstall),
			),
		).WithTheme(views.GetCustomTheme())

		err := form.Run()
		if err != nil {
			return fmt.Errorf("error prompting for Python installation: %w", err)
		}

		if confirmInstall {
			return installPython(hostname)
		} else {
			return errors.New("python3 is required but not installed")
		}
	}
	views.RenderInfoMessageBold("Python is already installed.")
	return nil
}

func installPython(hostname string) error {
	packageManager, err := detectPackageManager(hostname)
	if err != nil {
		return err
	}
	return installPythonWithPackageManager(hostname, packageManager)
}

// ensurePipInstalled checks if pip is installed and installs it if necessary
func ensurePipInstalled(hostname string, yesFlag bool) error {
	views.RenderInfoMessageBold("Checking pip installation...")

	// Check if pip is installed
	if err := runRemoteCommand(hostname, "python3 -m pip --version"); err != nil {
		if yesFlag {
			return installPip(hostname)
		}

		var confirmInstall bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("pip is not installed. Would you like to install it?").
					Description("pip is required to install Jupyter Notebook.").
					Value(&confirmInstall),
			),
		).WithTheme(views.GetCustomTheme())

		err := form.Run()
		if err != nil {
			return fmt.Errorf("error prompting for pip installation: %w", err)
		}

		if confirmInstall {
			return installPip(hostname)
		} else {
			return errors.New("pip is required but not installed")
		}
	}
	views.RenderInfoMessageBold("pip is installed.")
	return nil
}

func installPip(hostname string) error {
	// Check if we're on a Debian-based system
	if err := runRemoteCommand(hostname, "command -v apt-get"); err == nil {
		// We're on a Debian-based system, use apt to install pip and python3-venv
		views.RenderInfoMessageBold("Installing pip and python3-venv using apt...")
		if err := runRemoteCommand(hostname, "sudo apt-get update && sudo apt-get install -y python3-pip python3-venv"); err != nil {
			return fmt.Errorf("failed to install pip and python3-venv using apt: %w", err)
		}
	} else {
		// If not Debian-based, fall back to the get-pip.py method
		views.RenderInfoMessageBold("Installing pip using get-pip.py...")
		installCmd := "curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py && python3 get-pip.py --user && rm get-pip.py"
		if err := runRemoteCommand(hostname, installCmd); err != nil {
			return fmt.Errorf("failed to install pip: %w", err)
		}
	}
	return nil
}

// detectPackageManager detects the package manager on the remote host.
func detectPackageManager(hostname string) (string, error) {
	views.RenderInfoMessageBold("Detecting package manager...")
	commands := map[string]string{
		"apt-get": "dpkg -s apt >/dev/null 2>&1 && echo apt-get",
		"yum":     "rpm -q yum >/dev/null 2>&1 && echo yum",
		"brew":    "brew --version >/dev/null 2>&1 && echo brew",
	}
	for manager, cmd := range commands {
		if err := runRemoteCommand(hostname, cmd); err == nil {
			return manager, nil
		}
	}
	return "", errors.New("no supported package manager found")
}

// installPythonWithPackageManager installs Python using the detected package manager.
func installPythonWithPackageManager(hostname, manager string) error {
	views.RenderInfoMessageBold(fmt.Sprintf("Installing Python3 using %s...", manager))
	var installCmd string
	switch manager {
	case "apt-get":
		installCmd = "sudo apt-get update && sudo apt-get install -y python3"
	case "yum":
		installCmd = "sudo yum install -y python3"
	case "brew":
		installCmd = "brew install python3"
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
	return runRemoteCommand(hostname, installCmd)
}

// ensureJupyterInstalled checks if Jupyter Notebook is installed and installs it if necessary
func ensureJupyterInstalled(hostname string) error {
	views.RenderInfoMessageBold("Checking Jupyter Notebook installation...")

	// Check if Jupyter is installed
	checkCmd := ". ~/.jupyter_venv/bin/activate && jupyter --version"
	if err := runRemoteCommand(hostname, checkCmd); err == nil {
		views.RenderInfoMessageBold("Jupyter Notebook is already installed.")
		return nil
	}

	views.RenderInfoMessageBold("Installing python3-venv...")
	installVenvCmd := "sudo apt-get update && sudo apt-get install -y python3-venv"
	if err := runRemoteCommand(hostname, installVenvCmd); err != nil {
		return fmt.Errorf("failed to install python3-venv: %w", err)
	}

	views.RenderInfoMessageBold("Installing Jupyter Notebook in a virtual environment...")
	installCmd := `
		python3 -m venv ~/.jupyter_venv &&
		. ~/.jupyter_venv/bin/activate &&
		pip install notebook &&
		deactivate
	`
	return runRemoteCommand(hostname, installCmd)
}

// startJupyterServer starts the Jupyter Notebook server on the remote target.
func startJupyterServer(hostname string, activeProfile config.Profile, targetId, projectName string, gpgKey string) error {
	projectDir, err := util.GetProjectDir(activeProfile, targetId, projectName, gpgKey)
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("Starting Jupyter Notebook...")

	// Start Jupyter Notebook server in the background
	go func() {
		cmd := exec.CommandContext(context.Background(), "ssh", hostname, fmt.Sprintf(". ~/.jupyter_venv/bin/activate && cd %s && jupyter %s", projectDir, startJupyterCommand))
		cmd.Stdout = io.Writer(&util.DebugLogWriter{})
		cmd.Stderr = io.Writer(&util.DebugLogWriter{})
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	// Forward the IDE port
	browserPort, errChan := tailscale.ForwardPort(targetId, projectName, 8888, activeProfile)
	if browserPort == nil {
		if err := <-errChan; err != nil {
			return err
		}
	}

	// Open the browser with the forwarded port
	ideURL := fmt.Sprintf("http://localhost:%d", *browserPort)
	waitForPort(*browserPort)

	views.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s Jupyter Notebook port to %s.\nOpening browser...\n", projectName, ideURL))

	if err := browser.OpenURL(ideURL); err != nil {
		log.Error("Error opening URL: " + err.Error())
	}

	// Handle errors from forwarding
	for {
		err := <-errChan
		if err != nil {
			log.Debug(err)
		}
	}
}

// runRemoteCommand runs a command on the remote host and handles output.
func runRemoteCommand(hostname, command string) error {
	cmd := exec.Command("ssh", hostname, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

// waitForPort waits until the specified port is ready.
func waitForPort(port uint16) {
	for {
		if ports.IsPortReady(port) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}
