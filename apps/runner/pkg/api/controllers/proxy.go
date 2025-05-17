// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

var proxyTransport = &http.Transport{
	MaxIdleConns:        100,
	IdleConnTimeout:     90 * time.Second,
	MaxIdleConnsPerHost: 100,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
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
				if key != "Authorization" && key != "Cookie" {
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
func ProxyRequest(ctx *gin.Context) {
	sandboxId := ctx.Param("workspaceId")
	if sandboxId == "" {
		ctx.Error(common.NewBadRequestError(errors.New("sandbox ID is required")))
		return
	}

	runner := runner.GetInstance(nil)

	data := runner.SandboxService.GetSandboxStatesInfo(ctx, sandboxId)
	if data.SandboxState != enums.SandboxStateStarted {
		ctx.Error(common.NewBadRequestError(errors.New("sandbox is not started")))
		return
	}

	target, fullTargetURL, err := getProxyTarget(ctx, sandboxId)
	if err != nil {
		// Error already sent to the context
		return
	}

	fmt.Println(ctx.Param("path"))

	if regexp.MustCompile(`^/process/session/.+/command/.+/logs$`).MatchString(ctx.Param("path")) {
		if ctx.Query("follow") == "true" {
			ProxyCommandLogsStream(ctx, fullTargetURL)
			return
		}
	}

	// Create a new outgoing request
	outReq, err := http.NewRequestWithContext(
		ctx.Request.Context(),
		ctx.Request.Method,
		fullTargetURL,
		ctx.Request.Body,
	)
	if err != nil {
		ctx.Error(common.NewBadRequestError(fmt.Errorf("failed to create outgoing request: %w", err)))
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

	// Execute the request with our custom client that handles redirects
	resp, err := proxyClient.Do(outReq)
	if err != nil {
		ctx.Error(fmt.Errorf("proxy request failed: %w", err))
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

func getProxyTarget(ctx *gin.Context, sandboxId string) (*url.URL, string, error) {
	runner := runner.GetInstance(nil)

	targetURL, err := runner.Docker.GetContainerTargetURL(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.Error(err)
		return nil, "", err
	}

	err = runner.Docker.DaemonStartedCheck(ctx.Request.Context(), targetURL, 1, 1*time.Second, 0*time.Millisecond)
	if err != nil {
		startErr := runner.Docker.StartDaytonaDaemon(ctx.Request.Context(), sandboxId)
		if startErr != nil {
			ctx.Error(startErr)
			return nil, "", startErr
		}

		err = runner.Docker.DaemonStartedCheck(ctx.Request.Context(), targetURL, 10, 1*time.Second, 50*time.Millisecond)
		if err != nil {
			ctx.Error(err)
			return nil, "", err
		}
	}

	// Build the target URL
	target, err := url.Parse(targetURL)
	if err != nil {
		ctx.Error(common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, "", fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Get the wildcard path and normalize it
	path := ctx.Param("path")

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	fullTargetURL := fmt.Sprintf("%s%s", targetURL, path)
	if ctx.Request.URL.RawQuery != "" {
		fullTargetURL = fmt.Sprintf("%s?%s", fullTargetURL, ctx.Request.URL.RawQuery)
	}

	return target, fullTargetURL, nil
}
