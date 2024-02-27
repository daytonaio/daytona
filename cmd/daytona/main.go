// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/daytonaio/daytona/pkg/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
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

	cmd.Execute()
}
