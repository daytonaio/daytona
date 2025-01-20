// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"net/http"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api/controllers/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
					c.AbortWithError(http.StatusNotFound, errors.New("log file not found"))
					return
				}
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			defer logFile.Close()
			log.ReadLog(c, logFile, util.ReadLog, func(conn *websocket.Conn, messages chan []byte, errors chan error) {
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
				c.AbortWithError(http.StatusNotFound, errors.New("log file not found"))
				return
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		_, output := extractExitCode(string(content))
		c.String(http.StatusOK, output)
	}
}
