// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	golog "log"

	"github.com/daytonaio/daemon/cmd/daemon/config"
	"github.com/daytonaio/daemon/pkg/terminal"
	"github.com/daytonaio/daemon/pkg/toolbox"
	log "github.com/sirupsen/logrus"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	c.ProjectDir = filepath.Join(os.Getenv("HOME"))

	if projectDir := os.Getenv("DAYTONA_PROJECT_DIR"); projectDir != "" {
		c.ProjectDir = projectDir
	}

	if _, err := os.Stat(c.ProjectDir); os.IsNotExist(err) {
		if err := os.MkdirAll(c.ProjectDir, 0755); err != nil {
			panic(fmt.Errorf("failed to create project directory: %w", err))
		}
	}

	var logWriter io.Writer
	if c.LogFilePath != nil {
		logFile, err := os.OpenFile(*c.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error("Failed to open log file at ", *c.LogFilePath)
		} else {
			defer logFile.Close()
			logWriter = logFile
		}
	}

	initLogs(logWriter)

	errChan := make(chan error)

	toolBoxServer := &toolbox.Server{
		ProjectDir: c.ProjectDir,
	}

	// Start the toolbox server in a go routine
	go func() {
		err := toolBoxServer.Start()
		if err != nil {
			errChan <- err
		}
	}()

	// Start terminal server
	go func() {
		if err := terminal.StartTerminalServer(22222); err != nil {
			errChan <- err
		}
	}()

	err = <-errChan
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
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
