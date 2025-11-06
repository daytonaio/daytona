// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/toolbox_apiclient"

	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type PreviewLinkInput struct {
	SandboxId   string `json:"sandboxId" jsonschema:"ID of the sandbox to generate the preview link for."`
	Port        *int32 `json:"port" jsonschema:"Port to expose."`
	CheckServer *bool  `json:"checkServer,omitempty" jsonschema:"Check if a server is running on the specified port."`
	Description string `json:"description,omitempty" jsonschema:"Description of the service."`
}

type PreviewLinkOutput struct {
	PreviewURL string `json:"previewURL" jsonschema:"Preview URL of the service."`
	Accessible bool   `json:"accessible" jsonschema:"Whether the preview URL is accessible."`
	StatusCode string `json:"statusCode" jsonschema:"Status code of the preview URL."`
}

func (s *DaytonaSandboxMCPServer) getPreviewLinkTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "preview_link",
		Title:       "Preview Link",
		Description: "Generate accessible preview URLs for web applications running in the Daytona sandbox. Creates a secure tunnel to expose local ports externally without configuration. Validates if a server is actually running on the specified port and provides diagnostic information for troubleshooting. Supports custom descriptions and metadata for better organization of multiple services.",
	}
}

func (s *DaytonaSandboxMCPServer) handlePreviewLink(ctx context.Context, request *mcp.CallToolRequest, input *PreviewLinkInput) (*mcp.CallToolResult, *PreviewLinkOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.Port == nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("port parameter is required")
	}

	checkServer := false
	if input.CheckServer != nil && *input.CheckServer {
		checkServer = *input.CheckServer
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}
	defer stop()

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error extracting proxy URL: %v", err)
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DAYTONA_FS_MCP_SOURCE, sandbox.Id, proxyUrl, request.Extra.Header)

	log.Infof("Generating preview link - port: %d", *input.Port)

	// Check if server is running on specified port
	if checkServer {
		log.Infof("Checking if server is running - port: %d", *input.Port)

		checkCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' http://localhost:%d --max-time 2 || echo 'error'", *input.Port)
		result, _, err := toolboxApiClient.ProcessAPI.ExecuteCommand(ctx).Request(*toolbox_apiclient.NewExecuteRequest(checkCmd)).Execute()
		if err != nil {
			return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error checking server: %v", err)
		}

		response := strings.TrimSpace(result.Result)
		if response == "error" || strings.HasPrefix(response, "0") {
			log.Infof("No server detected - port: %d", *input.Port)

			// Check what might be using the port
			psCmd := fmt.Sprintf("ps aux | grep ':%d' | grep -v grep || echo 'No process found'", *input.Port)
			psResult, _, err := toolboxApiClient.ProcessAPI.ExecuteCommand(ctx).Request(*toolbox_apiclient.NewExecuteRequest(psCmd)).Execute()
			if err != nil {
				return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error checking processes: %v", err)
			}

			return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("no server detected on port %d. Process info: %s", *input.Port, strings.TrimSpace(psResult.Result))
		}
	}

	// Fetch preview URL
	previewURL, _, err := s.apiClient.SandboxAPI.GetPortPreviewUrl(ctx, input.SandboxId, float32(*input.Port)).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get preview URL: %v", err)
	}

	// Test URL accessibility if requested
	var accessible bool
	var statusCode string
	if checkServer {
		checkCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' %s --max-time 3 || echo 'error'", previewURL.Url)
		result, _, err := toolboxApiClient.ProcessAPI.ExecuteCommand(ctx).Request(*toolbox_apiclient.NewExecuteRequest(checkCmd)).Execute()
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

	log.Infof("Preview link generated: %s", previewURL.Url)
	log.Infof("Accessible: %t", accessible)
	log.Infof("Status code: %s", statusCode)

	return &mcp.CallToolResult{IsError: false}, &PreviewLinkOutput{
		PreviewURL: previewURL.Url,
		Accessible: accessible,
		StatusCode: statusCode,
	}, nil
}
