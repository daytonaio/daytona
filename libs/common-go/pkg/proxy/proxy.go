// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var proxyTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	TLSHandshakeTimeout: 10 * time.Second,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
}

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
			BufferPool:     proxyBufPool,
			ModifyResponse: modifyResponse,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Warnf("proxy error for %s%s: %v", r.Host, r.URL.Path, err)
				if !ctx.IsAborted() {
					ctx.AbortWithStatus(http.StatusBadGateway)
				}
			},
		}

		reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
