// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

const SandboxStateHeader = "X-Daytona-Sandbox-State"

var proxyTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	DialContext: (&net.Dialer{
		KeepAlive: 30 * time.Second,
	}).DialContext,
}

func NewProxyRequestHandler(getProxyTarget func(*gin.Context) (targetUrl *url.URL, extraHeaders map[string]string, err error), modifyResponse func(*http.Response) error, errorHandler func(http.ResponseWriter, *http.Request, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		target, extraHeaders, err := getProxyTarget(ctx)
		if err != nil {
			// Error already sent to the context
			return
		}

		if target == nil {
			return
		}

		reverseProxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.Host = target.Host
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = target.Path
				req.URL.RawPath = target.RawPath
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
		}

		reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
