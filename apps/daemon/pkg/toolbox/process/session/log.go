// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

func GetSessionCommandLogs(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")
		cmdId := c.Param("commandId")

		session, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
			return
		}

		command, ok := sessions[sessionId].commands[cmdId]
		if !ok {
			c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
			return
		}

		path := command.LogFilePath(session.Dir(configDir))

		if c.Request.Header.Get("Upgrade") == "websocket" {
			logFile, err := os.Open(path)
			if err != nil {
				if os.IsNotExist(err) {
					c.AbortWithError(http.StatusNotFound, err)
					return
				}
				if os.IsPermission(err) {
					c.AbortWithError(http.StatusForbidden, err)
					return
				}
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			defer logFile.Close()
			ReadLog(c, logFile, util.ReadLog, func(conn *websocket.Conn, messages chan []byte, errors chan error) {
				for {
					msg := <-messages
					_, output := extractExitCode(string(msg))
					err := conn.WriteMessage(websocket.TextMessage, []byte(output))
					if err != nil {
						errors <- err
						break
					}
				}
			})
			return
		}

		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				c.AbortWithError(http.StatusNotFound, err)
				return
			}
			if os.IsPermission(err) {
				c.AbortWithError(http.StatusForbidden, err)
				return
			}
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		_, output := extractExitCode(string(content))
		c.String(http.StatusOK, output)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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
