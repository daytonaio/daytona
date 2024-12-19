// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/logs"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ReadLogParams struct {
	Id                    string
	Label                 *string
	ServerUrl             string
	ApiKey                string
	SkipPrefixLengthSetup bool
	Index                 *int
	Follow                *bool
	Query                 *string
	From                  *time.Time
}

var targetLogsStarted bool

func ReadTargetLogs(ctx context.Context, params ReadLogParams) {
	checkAndSetupLongestPrefixLength(params.SkipPrefixLengthSetup, params.Id, params.Label)

	query := ""
	if params.Follow != nil && *params.Follow {
		query = "follow=true"
	}

	for {
		ws, res, err := util.GetWebsocketConn(ctx, fmt.Sprintf("/log/target/%s", params.Id), params.ServerUrl, params.ApiKey, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(apiclient.HandleErrorResponse(res, err))
			time.Sleep(250 * time.Millisecond)
			continue
		}

		readJSONLog(ctx, ws, logs_view.STATIC_INDEX, params.From)
		ws.Close()
		break
	}
}

func ReadWorkspaceLogs(ctx context.Context, params ReadLogParams) {
	checkAndSetupLongestPrefixLength(params.SkipPrefixLengthSetup, params.Id, params.Label)

	query := ""
	if params.Follow != nil && *params.Follow {
		query = "follow=true"
	}

	for {
		ws, res, err := util.GetWebsocketConn(ctx, fmt.Sprintf("/log/workspace/%s", params.Id), params.ServerUrl, params.ApiKey, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(apiclient.HandleErrorResponse(res, err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		index := 0
		if params.Index != nil {
			index = *params.Index
		}

		readJSONLog(ctx, ws, index, params.From)
		ws.Close()
		break
	}
}

func ReadBuildLogs(ctx context.Context, params ReadLogParams) {
	checkAndSetupLongestPrefixLength(params.SkipPrefixLengthSetup, params.Id, params.Label)

	for {
		var query string
		if params.Query != nil {
			query = *params.Query
		}

		ws, res, err := util.GetWebsocketConn(ctx, fmt.Sprintf("/log/build/%s", params.Id), params.ServerUrl, params.ApiKey, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(apiclient.HandleErrorResponse(res, err))
			time.Sleep(250 * time.Millisecond)
			continue
		}

		readJSONLog(ctx, ws, logs_view.FIRST_WORKSPACE_INDEX, nil)
		ws.Close()
		break
	}
}

func ReadRunnerLogs(ctx context.Context, params ReadLogParams) {
	checkAndSetupLongestPrefixLength(params.SkipPrefixLengthSetup, params.Id, params.Label)

	for {
		var query string
		if params.Query != nil {
			query = *params.Query
		}

		ws, res, err := util.GetWebsocketConn(ctx, fmt.Sprintf("/log/runner/%s", params.Id), params.ServerUrl, params.ApiKey, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(apiclient.HandleErrorResponse(res, err))
			time.Sleep(250 * time.Millisecond)
			continue
		}

		readJSONLog(ctx, ws, logs_view.FIRST_WORKSPACE_INDEX, nil)
		ws.Close()
		break
	}
}

func readJSONLog(ctx context.Context, ws *websocket.Conn, index int, from *time.Time) {
	logEntriesChan := make(chan logs.LogEntry)
	readErr := make(chan error)
	go func() {
		for {
			var logEntry logs.LogEntry

			err := ws.ReadJSON(&logEntry)

			// An empty entry will be sent from the server on close/EOF
			// We don't want to print that
			if logEntry != (logs.LogEntry{}) {
				logEntriesChan <- logEntry
			}

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log.Error(err)
				}
				readErr <- err
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case logEntry := <-logEntriesChan:
			if from != nil {
				parsedTime, err := time.Parse(time.RFC3339Nano, logEntry.Time)
				if err != nil {
					log.Trace(err)
				}

				if parsedTime.After(*from) || parsedTime.Equal(*from) {
					logs_view.DisplayLogEntry(logEntry, index)
				}
			} else {
				logs_view.DisplayLogEntry(logEntry, index)
			}

		case err := <-readErr:
			if err != nil {
				err := ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
				if err != nil {
					log.Trace(err)
				}
				ws.Close()
				return
			}
		}

		if !targetLogsStarted && index == logs_view.STATIC_INDEX {
			targetLogsStarted = true
		}
	}
}

func checkAndSetupLongestPrefixLength(skipSetup bool, id string, label *string) {
	if skipSetup {
		return
	}

	name := id
	if label != nil {
		name = *label
	}
	logs_view.SetupLongestPrefixLength([]string{name})
}
