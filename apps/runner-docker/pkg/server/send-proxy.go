// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	pb "github.com/daytonaio/runner/proto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProxyRequest handles all HTTP requests in a unified way
func (s *RunnerServer) ProxyRequest(ctx context.Context, req *pb.ProxyRequestMsg) (*pb.ProxyResponseMsg, error) {
	target, fullTargetURL, err := s.getProxyTarget(ctx, req.SandboxId, req.Path, req.QueryParams)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get proxy target: %v", err)
	}

	// Check if this is a streaming endpoint that should use ProxyStream instead
	if regexp.MustCompile(`^/process/session/.+/command/.+/logs$`).MatchString(req.Path) {
		if follow, exists := req.QueryParams["follow"]; exists && follow == "true" {
			return nil, status.Errorf(codes.InvalidArgument, "streaming endpoints should use ProxyStream method")
		}
	}

	// Create HTTP request exactly like the original Gin code
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullTargetURL, strings.NewReader(string(req.Body)))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create outgoing request: %v", err)
	}

	// Copy headers from original request (same logic as Gin)
	for key, value := range req.Headers {
		if key != "Connection" {
			httpReq.Header.Set(key, value)
		}
	}

	// Set the Host header to the target (same as Gin)
	httpReq.Host = target.Host
	httpReq.Header.Set("Connection", "keep-alive")

	// Execute the request with our custom client (same as Gin)
	resp, err := s.proxyClient.Do(httpReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response body: %v", err)
	}

	// Copy response headers (same as Gin)
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0] // Take first value for simplicity, could join multiple
		}
	}

	return &pb.ProxyResponseMsg{
		StatusCode: int32(resp.StatusCode),
		Headers:    headers,
		Body:       body,
	}, nil
}

// ProxyStream handles streaming requests (like command logs with follow=true)
func (s *RunnerServer) ProxyStream(req *pb.ProxyStreamRequest, stream pb.Runner_ProxyStreamServer) error {
	_, fullTargetURL, err := s.getProxyTarget(stream.Context(), req.SandboxId, req.Path, req.QueryParams)
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to get proxy target: %v", err)
	}

	// Convert HTTP URL to WebSocket URL (same as original ProxyCommandLogsStream)
	wsURL := strings.Replace(fullTargetURL, "http://", "ws://", 1)

	// Establish WebSocket connection
	ws, _, err := websocket.DefaultDialer.DialContext(stream.Context(), wsURL, nil)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create outgoing request: %v", err)
	}
	defer ws.Close()

	// Set close handler (same logic as Gin)
	ws.SetCloseHandler(func(code int, text string) error {
		closeMsg := &pb.ProxyStreamResponse{
			ResponseType: &pb.ProxyStreamResponse_Close{
				Close: &pb.StreamClose{
					Code:   int32(code),
					Reason: text,
				},
			},
		}
		stream.Send(closeMsg)
		return nil
	})

	// Stream messages (same logic as Gin but adapted for gRPC)
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return nil // Normal closure
				}
				log.Errorf("Error reading message: %v", err)

				// Send error to stream
				errorMsg := &pb.ProxyStreamResponse{
					ResponseType: &pb.ProxyStreamResponse_Error{
						Error: &pb.StreamError{
							Message: err.Error(),
							Code:    int32(codes.Internal),
						},
					},
				}
				stream.Send(errorMsg)
				return status.Errorf(codes.Internal, "websocket read error: %v", err)
			}

			// Send data chunk to stream
			dataMsg := &pb.ProxyStreamResponse{
				ResponseType: &pb.ProxyStreamResponse_Data{
					Data: &pb.StreamData{
						Content:   msg,
						Timestamp: time.Now().Unix(),
					},
				},
			}

			if err := stream.Send(dataMsg); err != nil {
				log.Errorf("Error writing message: %v", err)
				return status.Errorf(codes.Internal, "failed to send message: %v", err)
			}
		}
	}
}

// getProxyTarget builds the target URL (exact same logic as original Gin code)
func (s *RunnerServer) getProxyTarget(ctx context.Context, sandboxId, path string, queryParams map[string]string) (*url.URL, string, error) {
	// Get container details
	container, err := s.apiClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return nil, "", fmt.Errorf("sandbox container not found: %w", err)
	}

	var containerIP string
	for _, network := range container.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return nil, "", fmt.Errorf("container has no IP address, it might not be running")
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Ensure path always has a leading slash but not duplicate slashes (same as Gin)
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create the complete target URL with path
	fullTargetURL := fmt.Sprintf("%s%s", targetURL, path)

	// Add query parameters
	if len(queryParams) > 0 {
		values := url.Values{}
		for key, value := range queryParams {
			values.Add(key, value)
		}
		fullTargetURL = fmt.Sprintf("%s?%s", fullTargetURL, values.Encode())
	}

	return target, fullTargetURL, nil
}
