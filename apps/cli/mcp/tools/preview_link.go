// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func PreviewLink(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return nil, err
	}

	portStr := request.Params.Arguments["port"].(string)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	checkServer := request.Params.Arguments["check_server"].(bool)

	log.Infof("Generating preview link - port: %d", port)

	// Get sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err != nil || sandboxID == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox ID found in tracking file: %v", err)
	}

	// Get the sandbox using sandbox ID
	sandbox, _, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, sandboxID).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to get sandbox: %v", err)
	}

	if sandbox == nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox available")
	}

	// Check if server is running on specified port
	if checkServer {
		log.Infof("Checking if server is running - port: %d", port)

		checkCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://localhost:%d --max-time 2 || echo 'error'", port)
		result, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxID).ExecuteRequest(*daytonaapiclient.NewExecuteRequest(checkCmd)).Execute()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error checking server: %v", err)
		}

		response := strings.TrimSpace(result.Result)
		if response == "error" || strings.HasPrefix(response, "0") {
			log.Infof("No server detected - port: %d", port)

			// Check what might be using the port
			psCmd := fmt.Sprintf("ps aux | grep ':%d' | grep -v grep || echo 'No process found'", port)
			psResult, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxID).ExecuteRequest(*daytonaapiclient.NewExecuteRequest(psCmd)).Execute()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error checking processes: %v", err)
			}

			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no server detected on port %d. Process info: %s", port, strings.TrimSpace(psResult.Result))
		}
	}

	// Extract domain information from sandbox metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(*sandbox.Info.ProviderMetadata), &metadata); err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error parsing provider metadata: %v", err)
	}

	nodeDomain, ok := metadata["nodeDomain"].(string)
	if !ok {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("node domain not found in metadata")
	}

	// Format preview URL
	previewURL := fmt.Sprintf("http://%d-%s.%s", port, sandboxID, nodeDomain)

	// Test URL accessibility if requested
	var accessible bool
	var statusCode string
	if checkServer {
		checkCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' %s --max-time 3 || echo 'error'", previewURL)
		result, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxID).ExecuteRequest(*daytonaapiclient.NewExecuteRequest(checkCmd)).Execute()
		if err != nil {
			log.Errorf("Error checking preview URL: %v", err)
		} else {
			response := strings.TrimSpace(result.Result)
			accessible = response != "error" && !strings.HasPrefix(response, "0")
			if _, err := strconv.Atoi(response); err == nil {
				statusCode = response
			}
		}
	}

	log.Infof("Preview link generated: %s", previewURL)
	log.Infof("Accessible: %t", accessible)
	log.Infof("Status code: %s", statusCode)

	return mcp.NewToolResultText(fmt.Sprintf("Preview link generated: %s", previewURL)), nil
}
