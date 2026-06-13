// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// proxyTransport wraps *http.Transport with retryTransport for stale-connection
// handling. If direct *http.Transport access is needed, split into two variables.
//
// IdleConnTimeout must be shorter than the peer's idle timeout to avoid
// reusing a connection it already closed.
var proxyTransport http.RoundTripper = &retryTransport{
	base: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		TLSHandshakeTimeout: 10 * time.Second,
		IdleConnTimeout:     30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
}

// 32 KiB buffer pool shared across all proxied requests.
var proxyBufPool = &bufferPool{
	pool: sync.Pool{
		New: func() interface{} {
			b := make([]byte, 32*1024)
			return &b
		},
	},
}

type bufferPool struct {
	pool sync.Pool
}

func (bp *bufferPool) Get() []byte {
	buf := bp.pool.Get().(*[]byte)
	return *buf
}

func (bp *bufferPool) Put(b []byte) {
	bp.pool.Put(&b)
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
			Transport:  proxyTransport,
			BufferPool: proxyBufPool,
			// Periodic flushing covers fixed-Content-Length streams that Go's
			// auto-flush detection misses, without forcing a flush per write
			// on bulk transfers.
			FlushInterval:  100 * time.Millisecond,
			ModifyResponse: modifyResponse,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				// Client went away; nothing to report or write.
				if errors.Is(err, context.Canceled) {
					return
				}
				// Don't log r.Host: preview hosts can embed signed tokens.
				slog.Warn("proxy error", "method", r.Method, "path", r.URL.Path, "error", err)
				if ctx.Writer.Written() {
					ctx.Abort()
					return
				}
				ctx.AbortWithStatus(http.StatusBadGateway)
			},
		}

		reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
