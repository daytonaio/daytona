// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	golog "log"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/daytonaio/daytona/pkg/cmd/workspacemode"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "logs" {
		readCmdLogs(os.Args[2], os.Args[3])
		return
	}

	if internal.WorkspaceMode() {
		err := workspacemode.Execute()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
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

func readCmdLogs(sessionId, cmdId string) {
	ws, res, err := apiclient_util.GetWebsocketConn(context.Background(), fmt.Sprintf("/workspace/kubectl/kubectl/toolbox/process/session/%s/%s/logs", sessionId, cmdId), &config.Profile{
		Api: config.ServerApi{
			Url: "http://localhost:3986",
			Key: "OGY3ZDEyMDMtNzFmZi00ZDIxLWIzZDgtNWQ3OTk0ZjA2MWJk",
		},
	}, util.Pointer("follow=true"))
	if res.StatusCode == http.StatusNotFound {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				log.Fatal(err)
			} else {
				log.Info(err)
				return
			}
		}

		fmt.Println(string(msg))
	}
}
