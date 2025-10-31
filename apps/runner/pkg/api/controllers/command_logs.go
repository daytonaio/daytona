// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/proxy"
	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

func ProxyCommandLogsStream(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	targetURL, extraHeaders, err := getProxyTarget(ctx)
	if err != nil {
		// Error already sent to the context
		return
	}

	if ctx.Query("follow") != "true" {
		proxy.NewProxyRequestHandler(func(ctx *gin.Context) (*url.URL, map[string]string, error) {
			return targetURL, extraHeaders, nil
		})(ctx)
		return
	}

	fullTargetURL := strings.Replace(targetURL.String(), "http://", "ws://", 1)

	ws, _, err := websocket.DefaultDialer.DialContext(context.Background(), fullTargetURL+"?follow=true", nil)
	if err != nil {
		ctx.Error(errors.NewBadRequestError(fmt.Errorf("failed to create outgoing request: %w", err)))
		return
	}

	ctx.Header("Content-Type", "application/octet-stream")

	ws.SetCloseHandler(func(code int, text string) error {
		if code == websocket.CloseNormalClosure {
			return nil
		}
		ctx.AbortWithStatus(code)
		return nil
	})

	defer ws.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				slog.ErrorContext(reqCtx, "Error reading message", "error", err)
				ws.Close()
				return
			}

			_, err = ctx.Writer.Write(msg)
			if err != nil {
				slog.ErrorContext(reqCtx, "Error writing message", "error", err)
				ws.Close()
				return
			}
			ctx.Writer.Flush()
		}
	}
}
