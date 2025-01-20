// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"errors"
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

func writeJSONToWs(ws *websocket.Conn, c chan interface{}, errChan chan error) {
	for {
		err := ws.WriteJSON(<-c)
		if err != nil {
			errChan <- err
			break
		}
	}
}

// ReadLog reads from the logReader and writes to the websocket.
// T is the type of the message to be read from the logReader
func ReadLog[T any](ginCtx *gin.Context, logReader io.Reader, readFunc func(context.Context, io.Reader, bool, chan T, chan error), wsWriteFunc func(*websocket.Conn, chan T, chan error)) {
	followQuery := ginCtx.Query("follow")
	follow := followQuery == "true"

	ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		closeErr := websocket.CloseNormalClosure
		if !errors.Is(err, io.EOF) {
			closeErr = websocket.CloseInternalServerErr
		}
		err := ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeErr, ""), time.Now().Add(time.Second))
		if err != nil {
			log.Trace(err)
		}
		ws.Close()
	}()

	msgChannel := make(chan T)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(ginCtx.Request.Context())

	defer cancel()
	go readFunc(ctx, logReader, follow, msgChannel, errChannel)
	go wsWriteFunc(ws, msgChannel, errChannel)

	readErr := make(chan error)
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			readErr <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case err = <-errChannel:
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Error(err)
				}
				cancel()
				return
			}
		case err := <-readErr:
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				log.Error(err)
			}
			if err != nil {
				return
			}
		}
	}
}

func ReadServerLog(ginCtx *gin.Context) {
	s := server.GetInstance(nil)

	logFileQuery := ginCtx.Query("file")
	retryQuery := ginCtx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	if retry {
		for {
			reader, err := s.GetLogReader(logFileQuery)
			if err != nil && server.IsLogFileNotFound(err) {
				ginCtx.AbortWithError(http.StatusNotFound, err)
				return
			}
			if err == nil {
				ReadLog(ginCtx, reader, util.ReadLog, writeToWs)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	reader, err := s.GetLogReader(logFileQuery)
	if err != nil {
		if server.IsLogFileNotFound(err) {
			ginCtx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ginCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ReadLog(ginCtx, reader, util.ReadLog, writeToWs)
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
				ReadLog(ginCtx, wsLogReader, util.ReadJSONLog, writeJSONToWs)
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

	ReadLog(ginCtx, wsLogReader, util.ReadJSONLog, writeJSONToWs)
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
				ReadLog(ginCtx, projectLogReader, util.ReadJSONLog, writeJSONToWs)
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

	ReadLog(ginCtx, projectLogReader, util.ReadJSONLog, writeJSONToWs)
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
				ReadLog(ginCtx, buildLogReader, util.ReadJSONLog, writeJSONToWs)
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

	ReadLog(ginCtx, buildLogReader, util.ReadJSONLog, writeJSONToWs)
}
