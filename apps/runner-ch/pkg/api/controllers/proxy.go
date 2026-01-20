// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/runner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ProxyRequest handles proxying requests to a sandbox's container toolbox
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
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [get]
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [post]
//	@Router			/sandboxes/{sandboxId}/toolbox/{path} [delete]
func ProxyRequest(ctx *gin.Context) {
	// Check for command logs streaming request
	path := ctx.Param("path")
	if ShouldProxyCommandLogs(path) && ctx.Query("follow") == "true" {
		ProxyCommandLogsStream(ctx)
		return
	}

	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	target, err := getProxyTarget(ctx, r.CHClient)
	if err != nil {
		return // Error already set in context
	}

	// Use SSH tunnel for proxying since VMs are on remote CH host
	proxyWithSSHTunnel(ctx, r.CHClient, target)
}

// ProxyToPort handles proxying requests to applications on specific ports in the sandbox
//
//	@Tags			proxy
//	@Summary		Proxy requests to applications on specific ports
//	@Description	Forwards the request to the daemon's proxy endpoint which routes to localhost:port
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Param			port		path		string	true	"Target port number"
//	@Param			path		path		string	true	"Path to forward"
//	@Success		200			{object}	any		"Proxied response"
//	@Failure		400			{object}	string	"Bad request"
//	@Failure		401			{object}	string	"Unauthorized"
//	@Failure		404			{object}	string	"Sandbox container not found"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/proxy/{port}/{path} [get]
//	@Router			/sandboxes/{sandboxId}/proxy/{port}/{path} [post]
//	@Router			/sandboxes/{sandboxId}/proxy/{port}/{path} [delete]
func ProxyToPort(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	// For noVNC (port 6080), auto-add connection parameters to vnc.html requests
	port := ctx.Param("port")
	path := ctx.Param("path")
	if port == "6080" && strings.HasSuffix(path, "vnc.html") {
		if ctx.Query("autoconnect") == "" {
			filename := path
			if strings.HasPrefix(filename, "/") {
				filename = filename[1:]
			}
			redirectURL := filename + "?autoconnect=true&resize=scale"
			if ctx.Request.URL.RawQuery != "" {
				redirectURL = filename + "?" + ctx.Request.URL.RawQuery + "&autoconnect=true&resize=scale"
			}
			ctx.Header("Location", redirectURL)
			ctx.AbortWithStatus(http.StatusFound)
			return
		}
	}

	target, err := getProxyToPortTarget(ctx, r.CHClient)
	if err != nil {
		return // Error already set in context
	}

	// Use SSH tunnel for proxying since VMs are on remote CH host
	proxyWithSSHTunnel(ctx, r.CHClient, target)
}

// proxyWithSSHTunnel proxies requests through an SSH tunnel to the VM
func proxyWithSSHTunnel(ctx *gin.Context, client *cloudhypervisor.Client, target *url.URL) {
	transport := cloudhypervisor.GetSSHTunnelTransport(client.SSHHost, client.SSHKeyPath)

	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = target.Host
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.RawQuery = target.RawQuery

			// Forward the original query string if not set
			if req.URL.RawQuery == "" && ctx.Request.URL.RawQuery != "" {
				req.URL.RawQuery = ctx.Request.URL.RawQuery
			}

			log.Debugf("Proxying to %s", req.URL.String())
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Errorf("Proxy error: %v", err)
			http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		},
	}

	reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
}

// getProxyTarget builds the target URL for toolbox proxy requests
func getProxyTarget(ctx *gin.Context, client *cloudhypervisor.Client) (*url.URL, error) {
	sandboxId := ctx.Param("sandboxId")
	if sandboxId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "sandbox ID is required"})
		return nil, errors.New("sandbox ID is required")
	}

	// Get IP from cache or fetch it
	ipCache := cloudhypervisor.GetIPCache()
	containerIP := ipCache.GetOrFetch(ctx.Request.Context(), sandboxId, client)

	if containerIP == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not determine sandbox IP"})
		return nil, errors.New("could not determine sandbox IP")
	}

	// Build the target URL to daemon on port 2280
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)

	// Get the wildcard path and normalize it
	path := ctx.Param("path")
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse target URL: %v", err)})
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil
}

// getProxyToPortTarget builds the target URL for port-based proxy requests
func getProxyToPortTarget(ctx *gin.Context, client *cloudhypervisor.Client) (*url.URL, error) {
	sandboxId := ctx.Param("sandboxId")
	if sandboxId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "sandbox ID is required"})
		return nil, errors.New("sandbox ID is required")
	}

	port := ctx.Param("port")
	if port == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "port is required"})
		return nil, errors.New("port is required")
	}

	// Get IP from cache or fetch it
	ipCache := cloudhypervisor.GetIPCache()
	containerIP := ipCache.GetOrFetch(ctx.Request.Context(), sandboxId, client)

	if containerIP == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Could not determine sandbox IP"})
		return nil, errors.New("could not determine sandbox IP")
	}

	// Build the target URL to daemon's proxy endpoint
	// Daemon will forward /proxy/:port/* to localhost:port/*
	targetURL := fmt.Sprintf("http://%s:2280/proxy/%s", containerIP, port)

	// Get the wildcard path and normalize it
	path := ctx.Param("path")
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	target, err := url.Parse(fmt.Sprintf("%s%s", targetURL, path))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse target URL: %v", err)})
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil
}
