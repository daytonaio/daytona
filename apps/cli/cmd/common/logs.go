/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal/util"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ReadLogParams struct {
	Id           string
	ServerUrl    string
	ServerApi    config.ServerApi
	Follow       *bool
	ResourceType ResourceType
}

type ResourceType string

const (
	ResourceTypeWorkspace ResourceType = "workspace"
	ResourceTypeImage     ResourceType = "images"
)

func ReadBuildLogs(ctx context.Context, params ReadLogParams) {
	query := ""
	if params.Follow != nil && *params.Follow {
		query = "follow=true"
	}

	for {
		ws, res, err := util.GetWebsocketConn(ctx, fmt.Sprintf("/%s/%s/build-logs", params.ResourceType, params.Id), params.ServerUrl, params.ServerApi, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(apiclient.HandleErrorResponse(res, err))
			time.Sleep(250 * time.Millisecond)
			continue
		}

		readPlainTextLog(ctx, ws)
		ws.Close()
		break
	}
}

func readPlainTextLog(ctx context.Context, ws *websocket.Conn) {
	messagesChan := make(chan string)
	readErr := make(chan error)

	go func() {
		for {
			_, message, err := ws.ReadMessage()

			if len(message) > 0 {
				messagesChan <- string(message)
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
		case message := <-messagesChan:
			fmt.Println(message)
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
	}
}
