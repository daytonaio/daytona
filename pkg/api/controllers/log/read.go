// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/pkg/logs"
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

func ReadServerLog(ctx *gin.Context) {
	s := server.GetInstance(nil)

	logFileQuery := ctx.Query("file")
	retryQuery := ctx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	if retry {
		for {
			reader, err := s.GetLogReader(logFileQuery)
			if err != nil && server.IsLogFileNotFound(err) {
				ctx.AbortWithError(http.StatusNotFound, err)
				return
			}
			if err == nil {
				readLog(ctx, reader, logs.ReadLog, writeToWs)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	reader, err := s.GetLogReader(logFileQuery)
	if err != nil {
		if server.IsLogFileNotFound(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ctx, reader, logs.ReadLog, writeToWs)
}

func ReadTargetLog(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	retryQuery := ctx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			targetLogReader, err := server.TargetService.GetTargetLogReader(ctx.Request.Context(), targetId)
			if err == nil {
				readLog(ctx, targetLogReader, logs.ReadJSONLog, writeJSONToWs)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	targetLogReader, err := server.TargetService.GetTargetLogReader(ctx.Request.Context(), targetId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ctx, targetLogReader, logs.ReadJSONLog, writeJSONToWs)
}

func ReadWorkspaceLog(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	retryQuery := ctx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			workspaceLogReader, err := server.WorkspaceService.GetWorkspaceLogReader(ctx.Request.Context(), workspaceId)
			if err == nil {
				readLog(ctx, workspaceLogReader, logs.ReadJSONLog, writeJSONToWs)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	workspaceLogReader, err := server.WorkspaceService.GetWorkspaceLogReader(ctx.Request.Context(), workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ctx, workspaceLogReader, logs.ReadJSONLog, writeJSONToWs)
}

func ReadBuildLog(ctx *gin.Context) {
	buildId := ctx.Param("buildId")
	retryQuery := ctx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			buildLogReader, err := server.BuildService.GetBuildLogReader(ctx.Request.Context(), buildId)

			if err == nil {
				readLog(ctx, buildLogReader, logs.ReadJSONLog, writeJSONToWs)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	buildLogReader, err := server.BuildService.GetBuildLogReader(ctx.Request.Context(), buildId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ctx, buildLogReader, logs.ReadJSONLog, writeJSONToWs)
}

func ReadRunnerLog(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")
	retryQuery := ctx.DefaultQuery("retry", "true")
	retry := retryQuery == "true"

	server := server.GetInstance(nil)

	if retry {
		for {
			runnerLogReader, err := server.RunnerService.GetRunnerLogReader(ctx.Request.Context(), runnerId)
			if err == nil {
				readLog(ctx, runnerLogReader, logs.ReadJSONLog, writeJSONToWs)
				return
			}
			time.Sleep(TIMEOUT)
		}
	}

	runnerLogReader, err := server.RunnerService.GetRunnerLogReader(ctx.Request.Context(), runnerId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	readLog(ctx, runnerLogReader, logs.ReadJSONLog, writeJSONToWs)
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

// readLog reads from the logReader and writes to the websocket.
// T is the type of the message to be read from the logReader
func readLog[T any](ginCtx *gin.Context, logReader io.Reader, readFunc func(context.Context, io.Reader, bool, chan T, chan error), wsWriteFunc func(*websocket.Conn, chan T, chan error)) {
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
