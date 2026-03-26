// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/daytonaio/daytona/cli/config"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

type ExecuteRequest struct {
	Command string   `json:"command"`
	Cwd     *string  `json:"cwd,omitempty"`
	Timeout *float32 `json:"timeout,omitempty"`
}

type ExecuteResponse struct {
	ExitCode float32 `json:"exitCode"`
	Result   string  `json:"result"`
}

type Client struct {
	apiClient *apiclient.APIClient
}

func NewClient(apiClient *apiclient.APIClient) *Client {
	return &Client{
		apiClient: apiClient,
	}
}

// Gets the toolbox proxy URL for a sandbox, caching by region in config
func (c *Client) getProxyURL(ctx context.Context, sandboxId, region string) (string, error) {
	// Check config cache first
	cachedURL, err := config.GetToolboxProxyUrl(region)
	if err == nil && cachedURL != "" {
		return cachedURL, nil
	}

	// Fetch from API
	toolboxProxyUrl, _, err := c.apiClient.SandboxAPI.GetToolboxProxyUrl(ctx, sandboxId).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get toolbox proxy URL: %w", err)
	}

	// Best-effort caching
	_ = config.SetToolboxProxyUrl(region, toolboxProxyUrl.Url)

	return toolboxProxyUrl.Url, nil
}

func (c *Client) ExecuteCommand(ctx context.Context, sandbox *apiclient.Sandbox, request ExecuteRequest) (*ExecuteResponse, error) {
	proxyURL, err := c.getProxyURL(ctx, sandbox.Id, sandbox.Target)
	if err != nil {
		return nil, err
	}

	return c.executeCommandViaProxy(ctx, proxyURL, sandbox.Id, request)
}

// TODO: replace this with the toolbox api client at some point
func (c *Client) executeCommandViaProxy(ctx context.Context, proxyURL, sandboxId string, request ExecuteRequest) (*ExecuteResponse, error) {
	// Build the URL: {proxyUrl}/{sandboxId}/process/execute
	url := fmt.Sprintf("%s/%s/process/execute", strings.TrimSuffix(proxyURL, "/"), sandboxId)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return nil, err
	}

	if activeProfile.Api.Key != nil {
		req.Header.Set("Authorization", "Bearer "+*activeProfile.Api.Key)
	} else if activeProfile.Api.Token != nil {
		req.Header.Set("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)
	}

	if activeProfile.ActiveOrganizationId != nil {
		req.Header.Set("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var response ExecuteResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
