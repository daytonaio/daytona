// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	"github.com/daytonaio/runner-android/pkg/runner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ProxyRequest handles proxying requests to a sandbox's container toolbox
// Note: For Cuttlefish, we use ADB for most operations instead of HTTP proxy
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
	path := ctx.Param("path")

	// Handle computeruse/status - return active for Android (WebRTC is always available)
	if strings.HasSuffix(path, "computeruse/status") || path == "/computeruse/status" {
		handleComputerUseStatus(ctx)
		return
	}

	// Check for command logs streaming request
	if ShouldProxyCommandLogs(path) && ctx.Query("follow") == "true" {
		ProxyCommandLogsStream(ctx)
		return
	}

	// Route toolbox requests to ADB-based handlers
	HandleToolboxRequest(ctx)
}

// handleComputerUseStatus returns the status of computer use (WebRTC display) for Android
// For Cuttlefish, WebRTC is always available when the instance is running
func handleComputerUseStatus(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	// Check if the instance exists and is running
	instance, exists := r.CVDClient.GetInstance(sandboxId)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Sandbox not found"})
		return
	}

	// For Android/Cuttlefish, WebRTC is available when instance is running
	status := "inactive"
	if instance.State == cuttlefish.InstanceStateRunning {
		status = "active"
	} else if instance.State == cuttlefish.InstanceStateStarting {
		status = "partial"
	}

	log.Debugf("ComputerUse status for sandbox %s: %s (state: %s)", sandboxId, status, instance.State)
	ctx.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// ProxyToPort handles proxying requests to applications on specific ports in the sandbox
// For Android/Cuttlefish:
// - Port 6080: Proxies to Cuttlefish WebRTC operator for display streaming
// - Other ports: Advises using ADB port forwarding
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

	sandboxId := ctx.Param("sandboxId")
	port := ctx.Param("port")
	path := ctx.Param("path")

	// Get instance info
	instance, exists := r.CVDClient.GetInstance(sandboxId)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Sandbox not found"})
		return
	}

	// Special handling for port 6080 - WebRTC display proxy
	if port == "6080" {
		proxyWebRTC(ctx, r.CVDClient, instance, path)
		return
	}

	// For other ports, advise using ADB port forwarding
	info, err := r.CVDClient.GetSandboxInfo(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	log.Infof("Port proxy request for sandbox %s port %s (use ADB for Cuttlefish)", sandboxId, port)
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":     "Port proxy not supported for Cuttlefish. Use ADB port forwarding.",
		"adbSerial": info.ADBSerial,
		"hint":      fmt.Sprintf("adb -s %s forward tcp:<local_port> tcp:%s", info.ADBSerial, port),
	})
}

// proxyWebRTC handles proxying to the Cuttlefish WebRTC operator for display streaming
func proxyWebRTC(ctx *gin.Context, client *cuttlefish.Client, instance *cuttlefish.InstanceInfo, path string) {
	// Cuttlefish WebRTC operator runs on port 1443 (HTTPS)
	operatorPort := 1443

	// Get the actual device ID by querying the host orchestrator
	// Device ID format varies based on CVD group assignment: {group}-{instance}-{instance}
	deviceId := getDeviceIdForInstance(client, instance, operatorPort)
	if deviceId == "" {
		// Fallback to default pattern if lookup fails
		deviceId = fmt.Sprintf("cvd_1-%d-%d", instance.InstanceNum, instance.InstanceNum)
		log.Warnf("WebRTC proxy: could not look up device ID, using fallback: %s", deviceId)
	}
	deviceFilesPath := fmt.Sprintf("/devices/%s/files", deviceId)

	// Handle VNC-style URLs by redirecting to the correct Cuttlefish path structure
	// The WebRTC client JS parses location.pathname to extract device ID, so the
	// URL must be in the format /devices/{deviceId}/files/client.html
	if path == "" || path == "/" || path == "/vnc.html" || path == "vnc.html" {
		// Redirect to the correct path structure (browser will make new request)
		redirectPath := fmt.Sprintf("%s/client.html", deviceFilesPath)
		log.Infof("WebRTC proxy: redirecting to %s", redirectPath)
		ctx.Redirect(http.StatusFound, redirectPath)
		return
	}

	// Check if path is an operator API endpoint (pass through directly)
	// Known operator endpoints: /devices/, /infra_config, /connect, /forward, /poll_messages, /polled_connections
	isOperatorPath := strings.HasPrefix(path, "/devices/") ||
		strings.HasPrefix(path, "/infra_config") ||
		strings.HasPrefix(path, "/connect") ||
		strings.HasPrefix(path, "/forward") ||
		strings.HasPrefix(path, "/poll_messages") ||
		strings.HasPrefix(path, "/polled_connections")

	if !isOperatorPath {
		// Map relative asset requests (css, js, etc.) to the device's files directory
		if strings.HasPrefix(path, "/") {
			path = deviceFilesPath + path
		} else {
			path = deviceFilesPath + "/" + path
		}
		log.Debugf("WebRTC proxy: mapping asset request to %s", path)
	}

	// Build target URL
	var targetHost string
	if client.IsRemote() {
		// For remote mode, proxy through SSH tunnel
		targetHost = client.SSHHost
		// Extract just the host part (remove user@ if present)
		if idx := strings.Index(targetHost, "@"); idx != -1 {
			targetHost = targetHost[idx+1:]
		}
	} else {
		targetHost = "localhost"
	}

	targetURL := fmt.Sprintf("https://%s:%d", targetHost, operatorPort)
	target, err := url.Parse(targetURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Invalid target URL: %v", err)})
		return
	}

	log.Debugf("WebRTC proxy: forwarding %s to %s%s", ctx.Request.URL.Path, targetURL, path)

	// Check if this is a WebSocket upgrade request
	if isWebSocketRequest(ctx.Request) {
		proxyWebSocket(ctx, target, path)
		return
	}

	// Create reverse proxy for HTTP requests
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Configure TLS (Cuttlefish uses self-signed certs)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	// Modify the request
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = path
		req.Host = target.Host

		// Forward original query string
		if ctx.Request.URL.RawQuery != "" {
			req.URL.RawQuery = ctx.Request.URL.RawQuery
		}
	}

	// Handle errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Errorf("WebRTC proxy error: %v", err)
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
	}

	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// proxyWebSocket handles WebSocket connections for WebRTC signaling
func proxyWebSocket(ctx *gin.Context, target *url.URL, path string) {
	log.Debugf("WebSocket proxy: connecting to %s%s", target.Host, path)

	// Connect to target with TLS
	targetAddr := target.Host
	targetConn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		targetAddr,
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		log.Errorf("WebSocket proxy: failed to dial target: %v", err)
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Hijack the client connection
	hijacker, ok := ctx.Writer.(http.Hijacker)
	if !ok {
		log.Error("WebSocket proxy: response writer does not support hijacking")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Errorf("WebSocket proxy: failed to hijack connection: %v", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Build and send the WebSocket upgrade request to target
	req := ctx.Request
	requestLine := fmt.Sprintf("%s %s HTTP/1.1\r\n", req.Method, path)
	if _, err := targetConn.Write([]byte(requestLine)); err != nil {
		log.Errorf("WebSocket proxy: failed to write request line: %v", err)
		return
	}

	// Forward headers
	for key, values := range req.Header {
		for _, value := range values {
			header := fmt.Sprintf("%s: %s\r\n", key, value)
			if _, err := targetConn.Write([]byte(header)); err != nil {
				log.Errorf("WebSocket proxy: failed to write header: %v", err)
				return
			}
		}
	}
	// Add Host header
	hostHeader := fmt.Sprintf("Host: %s\r\n", target.Host)
	if _, err := targetConn.Write([]byte(hostHeader)); err != nil {
		log.Errorf("WebSocket proxy: failed to write host header: %v", err)
		return
	}
	if _, err := targetConn.Write([]byte("\r\n")); err != nil {
		log.Errorf("WebSocket proxy: failed to write header terminator: %v", err)
		return
	}

	// Bidirectional copy
	done := make(chan struct{})
	go func() {
		io.Copy(targetConn, clientConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, targetConn)
		done <- struct{}{}
	}()

	// Wait for either direction to finish
	<-done
	log.Debug("WebSocket proxy: connection closed")
}

// getProxyTarget builds the target URL for toolbox proxy requests
func getProxyTarget(ctx *gin.Context, client *cuttlefish.Client) (*url.URL, error) {
	sandboxId := ctx.Param("sandboxId")
	if sandboxId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "sandbox ID is required"})
		return nil, errors.New("sandbox ID is required")
	}

	// Get instance info
	info, err := client.GetSandboxInfo(ctx.Request.Context(), sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return nil, fmt.Errorf("sandbox not found: %w", err)
	}

	// For Cuttlefish, we'd use ADB serial, but return a placeholder URL
	targetURL := fmt.Sprintf("adb://%s", info.ADBSerial)

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

// deviceInfo represents a device from the host orchestrator
type deviceInfo struct {
	DeviceId  string `json:"device_id"`
	GroupName string `json:"group_name"`
	Name      string `json:"name"`
	ADBPort   int    `json:"adb_port"`
}

// getDeviceIdForInstance queries the host orchestrator to find the actual device ID
// CVD assigns group names dynamically, so we need to look up the device by ADB port
func getDeviceIdForInstance(client *cuttlefish.Client, instance *cuttlefish.InstanceInfo, operatorPort int) string {
	// Build target host
	var targetHost string
	if client.IsRemote() {
		targetHost = client.SSHHost
		if idx := strings.Index(targetHost, "@"); idx != -1 {
			targetHost = targetHost[idx+1:]
		}
	} else {
		targetHost = "localhost"
	}

	// Query the devices endpoint
	devicesURL := fmt.Sprintf("https://%s:%d/devices", targetHost, operatorPort)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Get(devicesURL)
	if err != nil {
		log.Debugf("WebRTC proxy: failed to query devices: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Debugf("WebRTC proxy: devices endpoint returned %d", resp.StatusCode)
		return ""
	}

	var devices []deviceInfo
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		log.Debugf("WebRTC proxy: failed to decode devices: %v", err)
		return ""
	}

	// Find device by ADB port
	for _, dev := range devices {
		if dev.ADBPort == instance.ADBPort {
			log.Debugf("WebRTC proxy: found device %s for instance %d (ADB port %d)", dev.DeviceId, instance.InstanceNum, instance.ADBPort)
			return dev.DeviceId
		}
	}

	// Fallback: find by instance name (the "name" field is the instance number as string)
	instanceName := fmt.Sprintf("%d", instance.InstanceNum)
	for _, dev := range devices {
		if dev.Name == instanceName {
			log.Debugf("WebRTC proxy: found device %s for instance %d by name", dev.DeviceId, instance.InstanceNum)
			return dev.DeviceId
		}
	}

	log.Debugf("WebRTC proxy: no device found for instance %d (ADB port %d)", instance.InstanceNum, instance.ADBPort)
	return ""
}
