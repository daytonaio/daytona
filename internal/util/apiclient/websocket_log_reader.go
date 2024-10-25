// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/logs"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var targetLogsStarted bool

func ReadTargetLogs(ctx context.Context, activeProfile config.Profile, targetId string, projectNames []string, follow, showTargetLogs bool, from *time.Time) {
	var wg sync.WaitGroup
	query := ""
	if follow {
		query = "follow=true"
	}

	if !showTargetLogs {
		targetLogsStarted = true
	}
	logs_view.CalculateLongestPrefixLength(projectNames)

	for index, projectName := range projectNames {
		wg.Add(1)
		go func(projectName string, from *time.Time) {
			defer wg.Done()

			for {
				// Make sure target logs started before showing any project logs
				if !targetLogsStarted {
					time.Sleep(250 * time.Millisecond)
					continue
				}

				ws, res, err := GetWebsocketConn(ctx, fmt.Sprintf("/log/target/%s/%s", targetId, projectName), &activeProfile, &query)
				// We want to retry getting the logs if it fails
				if err != nil {
					log.Trace(HandleErrorResponse(res, err))
					time.Sleep(500 * time.Millisecond)
					continue
				}

				readJSONLog(ctx, ws, index, from)
				ws.Close()
				break
			}
		}(projectName, from)
	}

	if showTargetLogs {
		for {
			ws, res, err := GetWebsocketConn(ctx, fmt.Sprintf("/log/target/%s", targetId), &activeProfile, &query)
			// We want to retry getting the logs if it fails
			if err != nil {
				log.Trace(HandleErrorResponse(res, err))
				time.Sleep(250 * time.Millisecond)
				continue
			}

			readJSONLog(ctx, ws, logs_view.STATIC_INDEX, from)
			ws.Close()
			break
		}
	}

	wg.Wait()
}

func ReadBuildLogs(ctx context.Context, activeProfile config.Profile, buildId string, query string) {
	logs_view.CalculateLongestPrefixLength([]string{buildId})

	for {
		ws, res, err := GetWebsocketConn(ctx, fmt.Sprintf("/log/build/%s", buildId), &activeProfile, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(HandleErrorResponse(res, err))
			time.Sleep(250 * time.Millisecond)
			continue
		}

		readJSONLog(ctx, ws, logs_view.FIRST_PROJECT_INDEX, nil)
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
