// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

var proxyTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	DialContext: (&net.Dialer{
		KeepAlive: 30 * time.Second,
	}).DialContext,
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Custom HTTP client that follows redirects while maintaining original headers
var proxyClient = &http.Client{
	Transport: proxyTransport,
	// Create a custom redirect policy
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// Copy headers from original request
		if len(via) > 0 {
			// Copy the headers from the original request
			for key, values := range via[0].Header {
				// Skip certain headers that shouldn't be copied
				if key != "Cookie" {
					for _, value := range values {
						req.Header.Add(key, value)
					}
				}
			}
		}

		// Limit the number of redirects to prevent infinite loops
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	},
}

// ProxyRequest handles proxying requests to a sandbox's container
//
//	@Tags			toolbox
//	@Summary		Proxy requests to the sandbox toolbox
//	@Description	Forwards the request to the specified sandbox's container
//	@Param			workspaceId	path		string	true	"Sandbox ID"
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			path		path		string	true	"Path to forward"
//	@Success		200			{object}	string	"Proxied response"
//	@Failure		400			{object}	string	"Bad request"
//	@Failure		401			{object}	string	"Unauthorized"
//	@Failure		404			{object}	string	"Sandbox container not found"
//	@Failure		409			{object}	string	"Sandbox container conflict"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/workspaces/{workspaceId}/{projectId}/toolbox/{path} [get]
func NewProxyRequestHandler(getProxyTarget func(*gin.Context) (*url.URL, string, map[string]string, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		target, fullTargetURL, extraHeaders, err := getProxyTarget(ctx)
		if err != nil {
			// Error already sent to the context
			return
		}

		// Create a new outgoing request
		outReq, err := http.NewRequestWithContext(
			ctx.Request.Context(),
			ctx.Request.Method,
			fullTargetURL,
			ctx.Request.Body,
		)
		if err != nil {
			ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to create outgoing request: %w", err)))
			return
		}

		// Copy headers from original request
		for key, values := range ctx.Request.Header {
			// Skip the Connection header
			if key != "Connection" {
				for _, value := range values {
					outReq.Header.Add(key, value)
				}
			}
		}

		// Set the Host header to the target
		outReq.Host = target.Host
		outReq.Header.Set("Connection", "keep-alive")

		for key, value := range extraHeaders {
			outReq.Header.Add(key, value)
		}

		if ctx.Request.Header.Get("Upgrade") == "websocket" {
			ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
			if err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			defer ws.Close()

			reqExtraHeaders := http.Header{}
			for key, value := range extraHeaders {
				reqExtraHeaders.Add(key, value)
			}

			conn, _, err := websocket.DefaultDialer.DialContext(ctx.Request.Context(), strings.Replace(fullTargetURL, "http", "ws", 1), reqExtraHeaders)
			if err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			defer conn.Close()

			go func() {
				io.Copy(ws.NetConn(), conn.NetConn())
			}()

			io.Copy(conn.NetConn(), ws.NetConn())

			return
		}

		// Execute the request with our custom client that handles redirects
		resp, err := proxyClient.Do(outReq)
		if err != nil {
			ctx.AbortWithError(http.StatusBadGateway, fmt.Errorf("proxy request failed: %w", err))
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				ctx.Writer.Header().Add(key, value)
			}
		}

		// Set the status code
		ctx.Writer.WriteHeader(resp.StatusCode)

		// Copy the response body
		if _, err := io.Copy(ctx.Writer, resp.Body); err != nil {
			log.Errorf("Error copying response body: %v", err)
			// Error already sent to client, just log here
		}
	}
}
