// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/logs"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/gorilla/websocket"
)

var workspaceLogsStarted bool

func ReadWorkspaceLogs(activeProfile config.Profile, workspaceId string, projectNames []string, stopLogs *bool) {
	var wg sync.WaitGroup
	query := "follow=true&retry=true"

	logs_view.CalculateLongestPrefixLength(projectNames)

	for index, projectName := range projectNames {
		wg.Add(1)
		go func(projectName string) {
			defer wg.Done()

			for {
				// Make sure workspace logs started before showing any project logs
				if !workspaceLogsStarted {
					time.Sleep(250 * time.Millisecond)
					continue
				}

				ws, _, err := GetWebsocketConn(fmt.Sprintf("/log/workspace/%s/%s", workspaceId, projectName), &activeProfile, &query)
				// We want to retry getting the logs if it fails
				if err != nil {
					// TODO: return log.Trace once https://github.com/daytonaio/daytona/issues/696 is resolved
					// log.Trace(apiclient_util.HandleErrorResponse(res, err))
					time.Sleep(500 * time.Millisecond)
					continue
				}

				readJSONLog(ws, stopLogs, index)
				ws.Close()
				break
			}
		}(projectName)
	}

	for {
		ws, _, err := GetWebsocketConn(fmt.Sprintf("/log/workspace/%s", workspaceId), &activeProfile, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			// TODO: return log.Trace once https://github.com/daytonaio/daytona/issues/696 is resolved
			// log.Trace(apiclient_util.HandleErrorResponse(res, err))
			time.Sleep(250 * time.Millisecond)
			continue
		}

		readJSONLog(ws, stopLogs, logs_view.WORKSPACE_INDEX)
		ws.Close()
		break
	}

	wg.Wait()
}

func readJSONLog(ws *websocket.Conn, stopLogs *bool, index int) {
	logEntriesChan := make(chan logs.LogEntry)
	go logs_view.DisplayLogs(logEntriesChan, index)

	for {
		var logEntry logs.LogEntry
		err := ws.ReadJSON(&logEntry)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		logEntriesChan <- logEntry

		if !workspaceLogsStarted && index == logs_view.WORKSPACE_INDEX {
			workspaceLogsStarted = true
		}

		if *stopLogs {
			return
		}
	}
}
