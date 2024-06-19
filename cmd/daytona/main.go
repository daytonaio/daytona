// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"time"
	"os/signal"
	"os/exec"
	"syscall"
	"fmt"

	golog "log"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/daytonaio/daytona/pkg/cmd/workspacemode"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

func main() {
	setupSignalHandler()

	if util.WorkspaceMode() {
		workspacemode.Execute()
		return
	}

	cmd.Execute()
}

func init() {
	logLevel := log.ErrorLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")
	if logLevelSet {
		switch logLevelEnv {
		case "debug":
			logLevel = log.DebugLevel
		case "info":
			logLevel = log.InfoLevel
		case "warn":
			logLevel = log.WarnLevel
		case "error":
			logLevel = log.ErrorLevel
		default:
			logLevel = log.ErrorLevel
		}
	}

	log.SetLevel(logLevel)

	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
		zerologLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(zerologLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
		Out:        &util.DebugLogWriter{},
		TimeFormat: time.RFC3339,
	})

	golog.SetOutput(&util.DebugLogWriter{})
}

// setupSignalHandler sets up a handler for SIGINT and SIGTERM to clean up the terminal.
func setupSignalHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cleanup()
		os.Exit(1)
	}()
}

// cleanup function to reset the terminal
func cleanup() {
	fmt.Println("\nAborting process and resetting terminal...")

	// Reset the terminal
	cmd := exec.Command("stty", "sane")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to reset terminal:", err)
	}

	fmt.Println("Terminal reset complete.")
}
