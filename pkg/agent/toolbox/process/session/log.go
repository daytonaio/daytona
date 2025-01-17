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

func SessionCommandLogs(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")
		cmdId := c.Param("commandId")

		if cmdId == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("commandId is required"))
			return
		}

		_, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(404, errors.New("session not found"))
			return
		}

		logFile, err := os.Open(filepath.Join(configDir, "sessions", sessionId, cmdId, "output.log"))
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
	}
}

// sudo wget -qO /usr/local/bin/websocat https://github.com/vi/websocat/releases/latest/download/websocat_max.aarch64-unknown-linux-musl
