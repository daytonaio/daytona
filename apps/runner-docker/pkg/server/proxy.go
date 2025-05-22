// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/runner-docker/internal/constants"
	"github.com/gorilla/websocket"

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

func (s *RunnerServer) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	authHeader := r.Header.Get(constants.AUTHORIZATION_HEADER)
	if authHeader == "" {
		http.Error(w, "authorization token is not provided", http.StatusUnauthorized)
		return
	}

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 || authParts[0] != constants.BEARER_AUTH_HEADER {
		http.Error(w, "invalid authorization token format", http.StatusUnauthorized)
		return
	}

	token := authParts[1]
	expectedToken := os.Getenv("TOKEN")

	if token != expectedToken {
		http.Error(w, "invalid authorization token", http.StatusUnauthorized)
		return
	}

	// Check if the endpoint matches the expected pattern
	if !regexp.MustCompile(`^/sandbox/[^/]+/toolbox/.*$`).MatchString(r.URL.Path) {
		http.Error(w, "invalid endpoint format", http.StatusBadRequest)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "invalid path format", http.StatusBadRequest)
		return
	}

	sandboxId := pathParts[2]
	pathArray := append([]string{""}, pathParts[4:]...)
	path := strings.Join(pathArray, "/")

	target, fullTargetURL, err := s.getProxyTarget(r.Context(), sandboxId, path, r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if regexp.MustCompile(`^/process/session/.+/command/.+/logs$`).MatchString(path) {
		if r.URL.Query().Get("follow") == "true" {
			proxyCommandLogsStream(w, r, fullTargetURL)
			return
		}
	}

	// Create a new outgoing request
	outReq, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		fullTargetURL,
		r.Body,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create outgoing request: %v", err), http.StatusBadRequest)
		return
	}

	// Copy headers from original request
	for key, values := range r.Header {
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
		http.Error(w, fmt.Sprintf("proxy request failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Errorf("Error copying response body: %v", err)
		// Error already sent to client, just log here
	}
}

func (s *RunnerServer) getProxyTarget(ctx context.Context, sandboxId string, path string, rawQuery string) (*url.URL, string, error) {
	// Get container details
	container, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return nil, "", fmt.Errorf("sandbox container not found: %w", err)
	}

	var containerIP string
	for _, network := range container.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return nil, "", errors.New("container has no IP address, it might not be running")
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	fullTargetURL := fmt.Sprintf("%s%s", targetURL, path)
	if rawQuery != "" {
		fullTargetURL = fmt.Sprintf("%s?%s", fullTargetURL, rawQuery)
	}

	return target, fullTargetURL, nil
}

func proxyCommandLogsStream(w http.ResponseWriter, r *http.Request, fullTargetURL string) {
	fullTargetURL = strings.Replace(fullTargetURL, "http://", "ws://", 1)

	ws, _, err := websocket.DefaultDialer.DialContext(r.Context(), fullTargetURL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create outgoing request: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	ws.SetCloseHandler(func(code int, text string) error {
		w.WriteHeader(code)
		return nil
	})

	defer ws.Close()

	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Errorf("Error reading message: %v", err)
				ws.Close()
				return
			}

			_, err = w.Write(msg)
			if err != nil {
				log.Errorf("Error writing message: %v", err)
				ws.Close()
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}()

	<-r.Context().Done()
}
