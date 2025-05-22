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
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/runner-docker/pkg/common"
	pb "github.com/daytonaio/runner/proto"
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

func (s *RunnerServer) SendProxy(ctx context.Context, req *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	target, fullTargetURL, err := s.getProxyTarget(ctx, req.GetSandboxId(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if regexp.MustCompile(`^/process/session/.+/command/.+/logs$`).MatchString(req.GetPath()) {
		if req.GetFollow() {
			return s.handleCommandLogsStream(ctx, fullTargetURL)
		}
	}

	// Create a new outgoing request
	outReq, err := http.NewRequestWithContext(
		ctx,
		req.GetMethod(),
		fullTargetURL,
		strings.NewReader(string(req.GetBody())),
	)
	if err != nil {
		return nil, common.NewBadRequestError(fmt.Errorf("failed to create outgoing request: %w", err))
	}

	// Copy headers from original request
	for key, value := range req.GetHeaders() {
		// Skip the Connection header
		if key != "Connection" {
			outReq.Header.Set(key, value)
		}
	}

	// Set the Host header to the target
	outReq.Host = target.Host
	outReq.Header.Set("Connection", "keep-alive")

	// Execute the request with our custom client that handles redirects
	resp, err := proxyClient.Do(outReq)
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert response headers to map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &pb.ProxyResponse{
		StatusCode: int32(resp.StatusCode),
		Headers:    headers,
		Body:       body,
	}, nil
}

func (s *RunnerServer) getProxyTarget(ctx context.Context, sandboxId, path string) (*url.URL, string, error) {
	// Get container details
	container, err := s.apiClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return nil, "", common.NewNotFoundError(fmt.Errorf("sandbox container not found: %w", err))
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
		return nil, "", common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	// Ensure path always has a leading slash but not duplicate slashes
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	fullTargetURL := fmt.Sprintf("%s%s", targetURL, path)

	return target, fullTargetURL, nil
}

func (s *RunnerServer) handleCommandLogsStream(ctx context.Context, fullTargetURL string) (*pb.ProxyResponse, error) {
	fullTargetURL = strings.Replace(fullTargetURL, "http://", "ws://", 1)

	ws, _, err := websocket.DefaultDialer.DialContext(ctx, fullTargetURL, nil)
	if err != nil {
		return nil, common.NewBadRequestError(fmt.Errorf("failed to create websocket connection: %w", err))
	}
	defer ws.Close()

	// Read all messages from the websocket
	var messages []string
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("Error reading websocket message: %v", err)
			}
			break
		}
		messages = append(messages, string(msg))
	}

	// Combine all messages into a single response
	combinedBody := strings.Join(messages, "\n")

	return &pb.ProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/octet-stream",
		},
		Body: []byte(combinedBody),
	}, nil
}
