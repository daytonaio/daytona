// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const TIMEOUT = 300 * time.Millisecond

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

func readLog(ginCtx *gin.Context, logReader io.Reader) {
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

func writeJSONToWs(ws *websocket.Conn, c chan interface{}, errChan chan error) {
	for {
		value := <-c
		err := ws.WriteJSON(value)
		if err != nil {
			errChan <- err
			break
		}
	}
}

func readJSONLog(ginCtx *gin.Context, logReader io.Reader) {
	followQuery := ginCtx.Query("follow")
	follow := followQuery == "true"
	retryQuery := ginCtx.Query("retry")
	retry := retryQuery == "true"

	ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer ws.Close()

	msgChannel := make(chan interface{})
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go util.ReadJSONLog(ctx, logReader, follow, retry, msgChannel, errChannel)
	go writeJSONToWs(ws, msgChannel, errChannel)

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
	retryQuery := ginCtx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	if retry {
		for {
			reader, err := server.GetLogReader()
			if err == nil {
				readLog(ginCtx, reader)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	reader, err := server.GetLogReader()
	if err != nil {
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ginCtx, reader)
}

func ReadWorkspaceLog(ginCtx *gin.Context) {
	workspaceId := ginCtx.Param("workspaceId")
	retryQuery := ginCtx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			wsLogReader, err := server.WorkspaceService.GetWorkspaceLogReader(workspaceId)
			if err == nil {
				readJSONLog(ginCtx, wsLogReader)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	wsLogReader, err := server.WorkspaceService.GetWorkspaceLogReader(workspaceId)
	if err != nil {
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readJSONLog(ginCtx, wsLogReader)
}

func ReadProjectLog(ginCtx *gin.Context) {
	workspaceId := ginCtx.Param("workspaceId")
	projectName := ginCtx.Param("projectName")
	retryQuery := ginCtx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			projectLogReader, err := server.WorkspaceService.GetProjectLogReader(workspaceId, projectName)
			if err == nil {
				readJSONLog(ginCtx, projectLogReader)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	projectLogReader, err := server.WorkspaceService.GetProjectLogReader(workspaceId, projectName)
	if err != nil {
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readJSONLog(ginCtx, projectLogReader)
}

func ReadBuildLog(ginCtx *gin.Context) {
	buildId := ginCtx.Param("buildId")
	retryQuery := ginCtx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			buildLogReader, err := server.BuildService.GetBuildLogReader(buildId)

			if err == nil {
				readJSONLog(ginCtx, buildLogReader)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	buildLogReader, err := server.BuildService.GetBuildLogReader(buildId)
	if err != nil {
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readJSONLog(ginCtx, buildLogReader)
}
