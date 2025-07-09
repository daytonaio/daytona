// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	proxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// ProxyRequest handles proxying requests to a sandbox's container
//
//	@Tags			toolbox
//	@Summary		Proxy requests to the sandbox toolbox
//	@Description	Forwards the request to the specified sandbox's container
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Param			path		path		string	true	"Path to forward"
//	@Success		200			{object}	any		"Proxied response"
//	@Failure		400			{object}	string	"Bad request"
//	@Failure		401			{object}	string	"Unauthorized"
//	@Failure		404			{object}	string	"Sandbox container not found"
//	@Failure		409			{object}	string	"Sandbox container conflict"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [get]
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [post]
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [delete]
func ProxyRequest(ctx *gin.Context) {
	if regexp.MustCompile(`^/process/session/.+/command/.+/logs$`).MatchString(ctx.Param("path")) {
		if ctx.Query("follow") == "true" {
			ProxyCommandLogsStream(ctx)
			return
		}
	}

	proxy.NewProxyRequestHandler(getProxyTarget)(ctx)
}

func getProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	runner := runner.GetInstance(nil)

	sandboxId := ctx.Param("sandboxId")
	if sandboxId == "" {
		ctx.Error(common.NewBadRequestError(errors.New("sandbox ID is required")))
		return nil, nil, errors.New("sandbox ID is required")
	}

	// Get container details
	container, err := runner.Docker.ContainerInspect(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.Error(common.NewNotFoundError(fmt.Errorf("sandbox container not found: %w", err)))
		return nil, nil, fmt.Errorf("sandbox container not found: %w", err)
	}

	var containerIP string
	for _, network := range container.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		ctx.Error(common.NewBadRequestError(errors.New("container has no IP address, it might not be running")))
		return nil, nil, errors.New("container has no IP address, it might not be running")
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)

	// Get the wildcard path and normalize it
	path := ctx.Param("path")

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		ctx.Error(common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil, nil
}
