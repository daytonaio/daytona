// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

var proxyTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	DisableCompression:  true,
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
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
func NewProxyRequestHandler(getProxyTarget func(*gin.Context) (targetUrl *url.URL, extraHeaders map[string]string, err error), modifyResponse func(*http.Response) error) gin.HandlerFunc {
	return NewProxyRequestHandlerWithErrorHandler(getProxyTarget, modifyResponse, nil)
}

// NewProxyRequestHandlerWithErrorHandler is like NewProxyRequestHandler but allows specifying a custom error handler
func NewProxyRequestHandlerWithErrorHandler(
	getProxyTarget func(*gin.Context) (targetUrl *url.URL, extraHeaders map[string]string, err error),
	modifyResponse func(*http.Response) error,
	errorHandler func(http.ResponseWriter, *http.Request, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTotal := time.Now()
		target, extraHeaders, err := getProxyTarget(ctx)
		getTargetTime := time.Since(startTotal)

		if err != nil {
			// Error already sent to the context
			return
		}

		reverseProxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.Host = target.Host
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = target.Path
				if target.RawQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
				}
				for key, value := range extraHeaders {
					req.Header.Add(key, value)
				}
			},
			Transport:      proxyTransport,
			ModifyResponse: modifyResponse,
			ErrorHandler:   errorHandler,
			FlushInterval:  -1, // Flush immediately for streaming
		}

		proxyStart := time.Now()
		reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
		proxyTime := time.Since(proxyStart)

		// Log timing for debugging slow requests (only if took > 1s)
		totalTime := time.Since(startTotal)
		if totalTime > 1*time.Second {
			fmt.Printf("[PROXY-SLOW] target=%s getTarget=%v proxy=%v total=%v\n", target.String(), getTargetTime, proxyTime, totalTime)
		}
	}
}
