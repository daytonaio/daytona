// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"time"

	golog "log"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

func main() {
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
