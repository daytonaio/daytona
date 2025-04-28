// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"fmt"
	"strings"

	"github.com/daytonaio/runner/pkg/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

func ProxyCommandLogsStream(ctx *gin.Context, fullTargetURL string) {
	fullTargetURL = strings.Replace(fullTargetURL, "http://", "ws://", 1)

	ws, _, err := websocket.DefaultDialer.DialContext(ctx, fullTargetURL, nil)
	if err != nil {
		ctx.Error(common.NewBadRequestError(fmt.Errorf("failed to create outgoing request: %w", err)))
		return
	}

	ctx.Header("Content-Type", "application/octet-stream")

	ws.SetCloseHandler(func(code int, text string) error {
		ctx.AbortWithStatus(code)
		return nil
	})

	defer ws.Close()

	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Errorf("Error reading message: %v", err)
				ws.Close()
				return
			}

			_, err = ctx.Writer.Write(msg)
			if err != nil {
				log.Errorf("Error writing message: %v", err)
				ws.Close()
				return
			}
			ctx.Writer.Flush()
		}
	}()

	<-ctx.Done()
}
