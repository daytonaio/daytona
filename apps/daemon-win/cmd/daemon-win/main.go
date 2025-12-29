// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"fmt"
	"io"
	golog "log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/daytonaio/daemon-win/cmd/daemon-win/config"
	"github.com/daytonaio/daemon-win/pkg/toolbox"
	log "github.com/sirupsen/logrus"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	var logWriter io.Writer
	if c.DaemonLogFilePath != "" {
		logFile, err := os.OpenFile(c.DaemonLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error("Failed to open log file at ", c.DaemonLogFilePath)
		} else {
			defer logFile.Close()
			logWriter = logFile
		}
	}

	initLogs(logWriter)

	// If workdir in image is not set, use user home as workdir
	if c.UserHomeAsWorkDir {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Warnf("failed to get home directory: %v", err)
		} else {
			err = os.Chdir(homeDir)
			if err != nil {
				log.Warnf("failed to change working directory to home directory: %v", err)
			}
		}
	}

	errChan := make(chan error)

	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get current working directory: %w", err))
	}

	// Ensure Windows Firewall allows daemon port
	ensureFirewallRule()

	toolBoxServer := &toolbox.Server{
		WorkDir: workDir,
	}

	// Start the toolbox server in a go routine
	go func() {
		err := toolBoxServer.Start()
		if err != nil {
			errChan <- err
		}
	}()

	// Set up signal handling for graceful shutdown
	// On Windows, we primarily handle os.Interrupt (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Wait for either an error or shutdown signal
	select {
	case err := <-errChan:
		log.Errorf("Error: %v", err)
	case sig := <-sigChan:
		log.Infof("Received signal %v, shutting down gracefully...", sig)
	}

	log.Info("Shutdown complete")
}

func initLogs(logWriter io.Writer) {
	logLevel := log.WarnLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")

	if logLevelSet {
		var err error
		logLevel, err = log.ParseLevel(logLevelEnv)
		if err != nil {
			logLevel = log.WarnLevel
		}
	}

	log.SetLevel(logLevel)
	logFormatter := &config.LogFormatter{
		TextFormatter: &log.TextFormatter{
			ForceColors: true,
		},
		LogFileWriter: logWriter,
	}

	log.SetFormatter(logFormatter)

	golog.SetOutput(log.New().WriterLevel(log.DebugLevel))
}

// ensureFirewallRule adds a Windows Firewall rule to allow incoming connections
// on the daemon port (2280). This is idempotent - it won't fail if the rule exists.
func ensureFirewallRule() {
	if runtime.GOOS != "windows" {
		return
	}

	// Use netsh to add firewall rule (works on all Windows versions)
	// The rule allows incoming TCP connections on port 2280
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=Daytona Daemon",
		"dir=in",
		"action=allow",
		"protocol=tcp",
		"localport=2280",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Error might mean rule already exists, which is fine
		log.Debugf("Firewall rule setup: %v (output: %s)", err, string(output))
	} else {
		log.Info("Windows Firewall rule added for Daytona Daemon (port 2280)")
	}
}
