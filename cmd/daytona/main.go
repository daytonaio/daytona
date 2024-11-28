// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"time"

	golog "log"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/daytonaio/daytona/pkg/cmd/workspacemode"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

func main() {
	if internal.WorkspaceMode() {
		err := workspacemode.Execute()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Add retry logic for workspace creation
    if os.Args[1] == "create" {
        err := executeWithRetry(cmd.Execute, 5, time.Second*2)
        if err != nil {
            log.Fatal(err)
        }
    } else {
        err := cmd.Execute()
        if err != nil {
            log.Fatal(err)
        }
    }
}

// executeWithRetry attempts to execute the given function with retries
func executeWithRetry(fn func() error, maxRetries int, delay time.Duration) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        if i < maxRetries-1 {
            log.Warnf("Error executing command: %v. Retrying in %v...", err, delay)
            time.Sleep(delay)
            delay *= 2 // Exponential backoff
        }
    }
    return fmt.Errorf("failed to execute command after %d attempts: %v", maxRetries, err)
}

func init() {
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
