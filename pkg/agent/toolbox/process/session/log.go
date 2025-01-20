// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api/controllers/log"
	"github.com/gin-gonic/gin"
)

func GetSessionCommandLogs(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")
		cmdId := c.Param("commandId")

		if cmdId == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("commandId is required"))
			return
		}

		_, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
			return
		}

		_, ok = sessions[sessionId].commands[cmdId]
		if !ok {
			c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
			return
		}

		path := filepath.Join(configDir, "sessions", sessionId, cmdId, "output.log")

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

			log.ReadLog(c, logFile, util.ReadLog, log.WriteToWs)
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

		c.String(http.StatusOK, string(content))
	}
}
