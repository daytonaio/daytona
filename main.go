// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/daytonaio/daytona/cli/cmd"
	"github.com/daytonaio/daytona/internal/util"

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
	log.SetFormatter(new(util.GRPCErrorFormatter))

	err := os.MkdirAll("/tmp/daytona", 0755)
	if err != nil {
		log.Fatal(err)
	}

	cmd.Execute()
}
