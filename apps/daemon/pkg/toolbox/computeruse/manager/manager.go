// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package manager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

type pluginRef struct {
	client *plugin.Client
	impl   computeruse.IComputerUse
	path   string
}

var ComputerUseHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_COMPUTER_USE_PLUGIN",
	MagicCookieValue: "daytona_computer_use",
}

var computerUse = &pluginRef{}

// ComputerUseError represents a computer-use plugin error with context
type ComputerUseError struct {
	Type    string // "dependency", "system", "plugin"
	Message string
	Details string
}

func (e *ComputerUseError) Error() string {
	return e.Message
}

// detectPluginError tries to execute the plugin binary directly to get detailed error information
func detectPluginError(path string) *ComputerUseError {
	// Try to execute the plugin directly to get error output
	cmd := exec.Command(path)

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Get the combined output
	output := stdout.String() + stderr.String()

	if err == nil {
		// Plugin executed successfully, this shouldn't happen in normal flow
		return &ComputerUseError{
			Type:    "plugin",
			Message: "Plugin executed successfully but failed during handshake",
			Details: "This may indicate a protocol version mismatch or plugin configuration issue.",
		}
	}

	// Get exit code if available
	exitCode := -1
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	// Log the raw error for debugging
	log.Debugf("Plugin execution failed - Exit code: %d, Error: %v", exitCode, err)
	log.Debugf("Plugin stdout: %s", stdout.String())
	log.Debugf("Plugin stderr: %s", stderr.String())

	// Check for missing X11 runtime dependencies
	if strings.Contains(output, "libX11.so.6") ||
		strings.Contains(output, "libXext.so.6") ||
		strings.Contains(output, "libXtst.so.6") ||
		strings.Contains(output, "libXrandr.so.2") ||
		strings.Contains(output, "libXrender.so.1") ||
		strings.Contains(output, "libXfixes.so.3") ||
		strings.Contains(output, "libXss.so.1") ||
		strings.Contains(output, "libXi.so.6") ||
		strings.Contains(output, "libXinerama.so.1") {

		missingLibs := []string{}
		if strings.Contains(output, "libX11.so.6") {
			missingLibs = append(missingLibs, "libX11")
		}
		if strings.Contains(output, "libXext.so.6") {
			missingLibs = append(missingLibs, "libXext")
		}
		if strings.Contains(output, "libXtst.so.6") {
			missingLibs = append(missingLibs, "libXtst")
		}
		if strings.Contains(output, "libXrandr.so.2") {
			missingLibs = append(missingLibs, "libXrandr")
		}
		if strings.Contains(output, "libXrender.so.1") {
			missingLibs = append(missingLibs, "libXrender")
		}
		if strings.Contains(output, "libXfixes.so.3") {
			missingLibs = append(missingLibs, "libXfixes")
		}
		if strings.Contains(output, "libXss.so.1") {
			missingLibs = append(missingLibs, "libXScrnSaver")
		}
		if strings.Contains(output, "libXi.so.6") {
			missingLibs = append(missingLibs, "libXi")
		}
		if strings.Contains(output, "libXinerama.so.1") {
			missingLibs = append(missingLibs, "libXinerama")
		}

		return &ComputerUseError{
			Type:    "dependency",
			Message: fmt.Sprintf("Computer-use plugin requires X11 runtime libraries that are not available (missing: %s)", strings.Join(missingLibs, ", ")),
			Details: fmt.Sprintf(`To enable computer-use functionality, install the required dependencies:

For Ubuntu/Debian:
  sudo apt-get update && sudo apt-get install -y \\
    libx11-6 libxrandr2 libxext6 libxrender1 libxfixes3 libxss1 libxtst6 libxi6 libxinerama1 \\
    xvfb x11vnc novnc xfce4 xfce4-terminal dbus-x11

For CentOS/RHEL/Fedora:
  sudo yum install -y libX11 libXrandr libXext libXrender libXfixes libXScrnSaver libXtst libXi libXinerama \\
    xorg-x11-server-Xvfb x11vnc novnc xfce4 xfce4-terminal dbus-x11

For Alpine:
  apk add --no-cache \\
    libx11 libxrandr libxext libxrender libxfixes libxss libxtst libxi libxinerama \\
    xvfb x11vnc novnc xfce4 xfce4-terminal dbus-x11

Raw error output: %s

Note: Computer-use features will be disabled until dependencies are installed.`, output),
		}
	}

	// Check for missing development libraries (build-time dependencies)
	if strings.Contains(output, "X11/extensions/XTest.h") ||
		strings.Contains(output, "X11/Xlib.h") ||
		strings.Contains(output, "X11/Xutil.h") ||
		strings.Contains(output, "X11/X.h") {

		return &ComputerUseError{
			Type:    "dependency",
			Message: "Computer-use plugin requires X11 development libraries",
			Details: fmt.Sprintf(`To build computer-use functionality, install the required development dependencies:

For Ubuntu/Debian:
  sudo apt-get update && sudo apt-get install -y \\
    libx11-dev libxtst-dev libxext-dev libxrandr-dev libxinerama-dev libxi-dev

For CentOS/RHEL/Fedora:
  sudo yum install -y libX11-devel libXtst-devel libXext-devel libXrandr-devel libXinerama-devel libXi-devel

For Alpine:
  apk add --no-cache \\
    libx11-dev libxtst-dev libxext-dev libxrandr-dev libxinerama-dev libxi-dev

Raw error output: %s

Note: Computer-use features will be disabled until dependencies are installed.`, output),
		}
	}

	// Check for permission issues
	if strings.Contains(output, "Permission denied") ||
		strings.Contains(output, "not executable") ||
		strings.Contains(output, "EACCES") {

		return &ComputerUseError{
			Type:    "system",
			Message: "Computer-use plugin has permission issues",
			Details: fmt.Sprintf("The plugin at %s is not executable. Please check file permissions and ensure the binary is executable.\n\nRaw error output: %s", path, output),
		}
	}

	// Check for architecture mismatch
	if strings.Contains(output, "wrong ELF class") ||
		strings.Contains(output, "architecture") ||
		strings.Contains(output, "ELF") ||
		strings.Contains(output, "exec format error") {

		return &ComputerUseError{
			Type:    "system",
			Message: "Computer-use plugin architecture mismatch",
			Details: fmt.Sprintf("The plugin was compiled for a different architecture. Please rebuild the plugin for the current system architecture.\n\nRaw error output: %s", output),
		}
	}

	// Check for missing system libraries
	if strings.Contains(output, "libc.so") ||
		strings.Contains(output, "libm.so") ||
		strings.Contains(output, "libdl.so") ||
		strings.Contains(output, "libpthread.so") {

		return &ComputerUseError{
			Type:    "system",
			Message: "Computer-use plugin requires basic system libraries",
			Details: fmt.Sprintf("The plugin is missing basic system libraries. This may indicate a corrupted binary or system issue.\n\nRaw error output: %s", output),
		}
	}

	// Check for file not found
	if strings.Contains(output, "No such file or directory") ||
		strings.Contains(output, "ENOENT") {

		return &ComputerUseError{
			Type:    "system",
			Message: "Computer-use plugin file not found",
			Details: fmt.Sprintf("The plugin file at %s could not be found or accessed.\n\nRaw error output: %s", path, output),
		}
	}

	// Check for Go runtime issues
	if strings.Contains(output, "go:") ||
		strings.Contains(output, "runtime:") ||
		strings.Contains(output, "panic:") {

		return &ComputerUseError{
			Type:    "plugin",
			Message: "Computer-use plugin has Go runtime issues",
			Details: fmt.Sprintf("The plugin encountered a Go runtime error.\n\nRaw error output: %s", output),
		}
	}

	// Generic plugin error with full details
	return &ComputerUseError{
		Type:    "plugin",
		Message: fmt.Sprintf("Computer-use plugin failed to start (exit code: %d)", exitCode),
		Details: fmt.Sprintf("Error: %v\nExit Code: %d\nOutput: %s", err, exitCode, output),
	}
}

func GetComputerUse(path string) (computeruse.IComputerUse, error) {
	if computerUse.impl != nil {
		return computerUse.impl, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Computer use plugin not found at %s. Skipping...", path)
		return nil, nil
	}

	pluginName := filepath.Base(path)
	pluginBasePath := filepath.Dir(path)

	if runtime.GOOS == "windows" && strings.HasSuffix(path, ".exe") {
		pluginName = strings.TrimSuffix(pluginName, ".exe")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name: pluginName,
		// Output: log.New().WriterLevel(log.DebugLevel),
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &computeruse.ComputerUsePlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: ComputerUseHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(path),
		Logger:          logger,
		Managed:         true,
	})

	log.Infof("Computer use %s registered", pluginName)

	rpcClient, err := client.Client()
	if err != nil {
		// Try to get detailed error information
		pluginErr := detectPluginError(path)

		switch pluginErr.Type {
		case "dependency":
			log.Warn(pluginErr.Message)
			log.Info(pluginErr.Details)
			log.Info("Continuing without computer-use functionality...")
			return nil, nil // Return nil to continue without the plugin

		case "system":
			log.Error(pluginErr.Message)
			log.Info(pluginErr.Details)
			log.Info("Continuing without computer-use functionality...")
			return nil, nil // Return nil to continue without the plugin

		default:
			log.Error(pluginErr.Message)
			log.Info(pluginErr.Details)
			log.Info("Continuing without computer-use functionality...")
			return nil, nil // Return nil to continue without the plugin
		}
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		log.Errorf("Failed to dispense computer-use plugin: %v", err)
		log.Info("Continuing without computer-use functionality...")
		return nil, nil
	}

	impl, ok := raw.(computeruse.IComputerUse)
	if !ok {
		log.Errorf("Unexpected type from computer-use plugin")
		log.Info("Continuing without computer-use functionality...")
		return nil, nil
	}

	_, err = impl.Initialize()
	if err != nil {
		log.Errorf("Failed to initialize computer-use plugin: %v", err)
		log.Info("Continuing without computer-use functionality...")
		return nil, nil
	}

	log.Info("Computer-use plugin initialized successfully")
	computerUse.client = client
	computerUse.impl = impl
	computerUse.path = pluginBasePath

	return impl, nil
}
