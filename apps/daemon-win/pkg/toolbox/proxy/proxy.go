// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/daytonaio/daemon-win/pkg/common"
	"github.com/gin-gonic/gin"
)

// ProxyHandler handles proxy requests to localhost ports
func ProxyHandler(c *gin.Context) {
	targetPort := c.Param("port")
	if targetPort == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, common.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "target port is required",
			Code:       "BAD_REQUEST",
			Timestamp:  time.Now(),
			Path:       c.Request.URL.Path,
			Method:     c.Request.Method,
		})
		return
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://localhost:%s", targetPort)

	// Get the wildcard path and normalize it
	path := c.Param("path")

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, common.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("failed to parse target URL: %v", err),
			Code:       "BAD_REQUEST",
			Timestamp:  time.Now(),
			Path:       c.Request.URL.Path,
			Method:     c.Request.Method,
		})
		return
	}

	// Copy query string
	target.RawQuery = c.Request.URL.RawQuery

	// Check if this is a WebSocket upgrade request
	isWebSocket := c.Request.Header.Get("Upgrade") == "websocket"

	// Create reverse proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = target
			req.Host = target.Host
			req.Header = c.Request.Header.Clone()

			// Remove hop-by-hop headers, but preserve WebSocket headers for upgrades
			headersToRemove := []string{"Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailers", "Transfer-Encoding"}
			if !isWebSocket {
				// Only remove Connection and Upgrade for non-WebSocket requests
				headersToRemove = append(headersToRemove, "Connection", "Upgrade")
			}
			for _, h := range headersToRemove {
				req.Header.Del(h)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if errors.Is(err, io.EOF) {
				return
			}
			c.AbortWithStatusJSON(http.StatusBadGateway, common.ErrorResponse{
				StatusCode: http.StatusBadGateway,
				Message:    fmt.Sprintf("proxy error: %v", err),
				Code:       "BAD_GATEWAY",
				Timestamp:  time.Now(),
				Path:       c.Request.URL.Path,
				Method:     c.Request.Method,
			})
		},
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
