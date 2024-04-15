// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func writeToWs(ws *websocket.Conn, c chan []byte, errChan chan error) {
	for {
		err := ws.WriteMessage(websocket.TextMessage, <-c)
		if err != nil {
			errChan <- err
			break
		}
	}
}

func readLog(ginCtx *gin.Context, logReader *io.Reader) {
	followQuery := ginCtx.Query("follow")
	follow := followQuery == "true"

	ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer ws.Close()

	msgChannel := make(chan []byte)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go util.ReadLog(ctx, logReader, follow, msgChannel, errChannel)
	go writeToWs(ws, msgChannel, errChannel)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := <-errChannel
				if err != nil {
					if err.Error() != "EOF" {
						log.Error(err)
					}
					ws.Close()
					cancel()
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := ws.ReadMessage()
			if err != nil {
				ws.Close()
				cancel()
				return
			}
		}
	}
}

func ReadServerLog(ginCtx *gin.Context) {
	server := server.GetInstance(nil)

	reader, err := server.GetLogReader()
	if err != nil {
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ginCtx, &reader)
}

func ReadWorkspaceLog(ginCtx *gin.Context) {
	workspaceId := ginCtx.Param("workspaceId")

	server := server.GetInstance(nil)

	wsLogReader, err := server.WorkspaceService.GetWorkspaceLogReader(workspaceId)
	if err != nil {
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ginCtx, &wsLogReader)
}
