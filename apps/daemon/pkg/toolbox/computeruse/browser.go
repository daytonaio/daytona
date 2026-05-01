// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	commonproxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func BrowserProxyPath(localURL string) string {
	u, err := url.Parse(localURL)
	if err != nil || u.Path == "" {
		return ""
	}
	return "/browser/cdp" + u.EscapedPath()
}

func RewriteBrowserCDPURL(localURL, externalBaseURL string) string {
	proxyPath := BrowserProxyPath(localURL)
	if proxyPath == "" || externalBaseURL == "" {
		return localURL
	}

	base, err := url.Parse(strings.TrimRight(externalBaseURL, "/"))
	if err != nil || base.Host == "" {
		return localURL
	}

	if base.Scheme == "http" {
		base.Scheme = "ws"
	} else {
		base.Scheme = "wss"
	}
	base.Path = strings.TrimRight(base.EscapedPath(), "/") + proxyPath
	base.RawQuery = ""
	base.Fragment = ""
	return base.String()
}

func externalBaseURL(r *http.Request) string {
	if base := firstHeader(r, commonproxy.DaytonaToolboxBaseURLHeader); base != "" {
		return strings.TrimRight(base, "/")
	}

	host := firstHeader(r, "X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	if host == "" {
		return ""
	}

	proto := firstHeader(r, "X-Forwarded-Proto")
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	prefix := strings.TrimRight(firstHeader(r, "X-Forwarded-Prefix"), "/")
	return proto + "://" + host + prefix
}

func firstHeader(r *http.Request, key string) string {
	value := r.Header.Get(key)
	if i := strings.Index(value, ","); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}

func rewriteBrowserCDPResponse(r *http.Request, response *BrowserCDPResponse) {
	if response == nil {
		return
	}
	if response.ProxyPath == "" {
		response.ProxyPath = BrowserProxyPath(response.LocalWebSocketDebuggerURL)
	}
	if response.WebSocketDebuggerURL == "" || isLocalBrowserCDPURL(response.WebSocketDebuggerURL) {
		response.WebSocketDebuggerURL = RewriteBrowserCDPURL(response.LocalWebSocketDebuggerURL, externalBaseURL(r))
	}
}

func rewriteBrowserStatusResponse(r *http.Request, response *BrowserStatusResponse) {
	if response == nil {
		return
	}
	if response.ProxyPath == "" {
		response.ProxyPath = BrowserProxyPath(response.LocalWebSocketDebuggerURL)
	}
	if response.WebSocketDebuggerURL == "" || isLocalBrowserCDPURL(response.WebSocketDebuggerURL) {
		response.WebSocketDebuggerURL = RewriteBrowserCDPURL(response.LocalWebSocketDebuggerURL, externalBaseURL(r))
	}
}

func isLocalBrowserCDPURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil {
		return false
	}
	switch u.Hostname() {
	case "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}

// GetBrowserCDP godoc
//
//	@Summary		Get browser CDP URL
//	@Description	Lazily start managed Chromium and return a CDP WebSocket URL
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	BrowserCDPResponse
//	@Router			/computeruse/browser/cdp [get]
//
//	@id				GetBrowserCDP
func WrapBrowserCDPHandler(fn func(*BrowserCDPRequest) (*BrowserCDPResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn(&BrowserCDPRequest{ExternalBaseURL: externalBaseURL(c.Request)})
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		rewriteBrowserCDPResponse(c.Request, response)
		c.JSON(http.StatusOK, response)
	}
}

// GetBrowserStatus godoc
//
//	@Summary		Get browser status
//	@Description	Get the managed Chromium process status
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	BrowserStatusResponse
//	@Router			/computeruse/browser/status [get]
//
//	@id				GetBrowserStatus
func WrapBrowserStatusHandler(fn func() (*BrowserStatusResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		rewriteBrowserStatusResponse(c.Request, response)
		c.JSON(http.StatusOK, response)
	}
}

// StopBrowser godoc
//
//	@Summary		Stop browser
//	@Description	Stop the managed Chromium process
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	Empty
//	@Router			/computeruse/browser/stop [post]
//
//	@id				StopBrowser
func WrapStopBrowserHandler(fn func() (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapBrowserCDPProxyHandler(status func() (*BrowserStatusResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := status()
		if err != nil || response == nil || !response.Running || response.Port == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "browser is not running"})
			return
		}

		target := &url.URL{Scheme: "http", Host: "127.0.0.1:" + strconv.Itoa(response.Port)}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = c.Param("path")
			req.URL.RawPath = ""
			req.Host = target.Host
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
