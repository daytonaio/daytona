// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type PreviewLinkArgs struct {
	Id          *string `json:"id,omitempty"`
	Port        *int32  `json:"port,omitempty"`
	CheckServer *bool   `json:"checkServer,omitempty"`
	Description *string `json:"description,omitempty"`
}

func GetPreviewLinkTool() mcp.Tool {
	return mcp.NewTool("preview_link",
		mcp.WithDescription("Generate accessible preview URLs for web applications running in the Daytona sandbox. Creates a secure tunnel to expose local ports externally without configuration. Validates if a server is actually running on the specified port and provides diagnostic information for troubleshooting. Supports custom descriptions and metadata for better organization of multiple services."),
		mcp.WithNumber("port", mcp.Required(), mcp.Description("Port to expose.")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Description of the service.")),
		mcp.WithBoolean("checkServer", mcp.Required(), mcp.Description("Check if a server is running on the specified port.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to generate the preview link for.")),
	)
}

func PreviewLink(ctx context.Context, request mcp.CallToolRequest, args PreviewLinkArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return nil, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	if args.Port == nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("port parameter is required")
	}

	checkServer := false
	if args.CheckServer != nil && *args.CheckServer {
		checkServer = *args.CheckServer
	}

	log.Infof("Generating preview link - port: %d", *args.Port)

	// Get the sandbox using sandbox ID
	sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, *args.Id).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to get sandbox: %v", err)
	}

	if sandbox == nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox available")
	}

	// Check if server is running on specified port
	if checkServer {
		log.Infof("Checking if server is running - port: %d", *args.Port)

		checkCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://localhost:%d --max-time 2 || echo 'error'", *args.Port)
		result, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, *args.Id).ExecuteRequest(*apiclient.NewExecuteRequest(checkCmd)).Execute()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error checking server: %v", err)
		}

		response := strings.TrimSpace(result.Result)
		if response == "error" || strings.HasPrefix(response, "0") {
			log.Infof("No server detected - port: %d", *args.Port)

			// Check what might be using the port
			psCmd := fmt.Sprintf("ps aux | grep ':%d' | grep -v grep || echo 'No process found'", *args.Port)
			psResult, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, *args.Id).ExecuteRequest(*apiclient.NewExecuteRequest(psCmd)).Execute()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error checking processes: %v", err)
			}

			return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no server detected on port %d. Process info: %s", *args.Port, strings.TrimSpace(psResult.Result))
		}
	}

	var runnerDomain string
	if sandbox.RunnerDomain != nil {
		runnerDomain = *sandbox.RunnerDomain
	}

	// Format preview URL
	previewURL := fmt.Sprintf("http://%d-%s.%s", *args.Port, *args.Id, runnerDomain)

	// Test URL accessibility if requested
	var accessible bool
	var statusCode string
	if checkServer {
		checkCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' %s --max-time 3 || echo 'error'", previewURL)
		result, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, *args.Id).ExecuteRequest(*apiclient.NewExecuteRequest(checkCmd)).Execute()
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
