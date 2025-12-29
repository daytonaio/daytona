// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	proxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/daytonaio/runner-win/pkg/libvirt"
	"github.com/daytonaio/runner-win/pkg/runner"
	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
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

	r := runner.GetInstance(nil)

	// Dev environment: use SSH tunnel for remote libvirt
	if libvirt.ShouldUseSSHTunnel(r.LibVirt.GetURI()) {
		proxyWithSSHTunnel(ctx, r.LibVirt)
		return
	}

	proxy.NewProxyRequestHandler(getProxyTarget, nil)(ctx)
}

// proxyWithSSHTunnel proxies requests through an SSH tunnel for dev environments
// where libvirt runs on a remote machine
func proxyWithSSHTunnel(ctx *gin.Context, lv *libvirt.LibVirt) {
	target, _, err := getProxyTarget(ctx)
	if err != nil {
		return
	}

	sshHost := lv.GetSSHHost()
	transport := libvirt.GetSSHTunnelTransport(sshHost)

	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = target.Host
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.RawQuery = target.RawQuery
		},
		Transport: transport,
	}

	reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
}

func getProxyTarget(ctx *gin.Context) (*url.URL, map[string]string, error) {
	runner := runner.GetInstance(nil)

	sandboxId := ctx.Param("sandboxId")
	if sandboxId == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("sandbox ID is required")))
		return nil, nil, errors.New("sandbox ID is required")
	}

	// Get container details
	container, err := runner.LibVirt.ContainerInspect(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.Error(common_errors.NewNotFoundError(fmt.Errorf("sandbox container not found: %w", err)))
		return nil, nil, fmt.Errorf("sandbox container not found: %w", err)
	}

	containerIP := libvirt.GetDomainIpAddress(ctx.Request.Context(), container)

	if containerIP == "" {
		message := "no IP address found. Is the Sandbox started?"
		ctx.Error(common_errors.NewBadRequestError(errors.New(message)))
		return nil, nil, errors.New(message)
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
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err)))
		return nil, nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil, nil
}
