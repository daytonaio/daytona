// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
}

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

var TokenNewCmd = &cobra.Command{
	Use:     "token",
	Short:   "Create a new API token via browser authentication",
	Long:    "Opens a browser window for authentication and automatically saves the API token to your config",
	Args:    cobra.NoArgs,
	GroupID: internal.USER_GROUP,
	RunE:    runTokenNew,
}

var (
	profileFlag string
)

func init() {
	TokenNewCmd.Flags().StringVar(&profileFlag, "profile", "", "Profile name to save token under")
}

func runTokenNew(cmd *cobra.Command, args []string) error {
	apiURL := config.GetDaytonaApiUrl()

	if internal.Version == "v0.0.0-dev" {
		apiURL = "http://localhost:3001/api"
	}

	// Step 1: Request device code
	view_common.RenderInfoMessage("Requesting device authorization...")

	deviceResp, err := requestDeviceCode(apiURL)
	if err != nil {
		return fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Display code and open browser
	fmt.Println()
	view_common.RenderInfoMessageBold("Please visit this URL in your browser:")
	fmt.Println()
	fmt.Printf("    %s\n", deviceResp.VerificationURIComplete)
	fmt.Println()
	view_common.RenderInfoMessage("Or go to " + deviceResp.VerificationURI + " and enter code:")
	fmt.Println()
	fmt.Printf("    %s\n", deviceResp.UserCode)
	fmt.Println()

	// Try to open browser automatically
	if err := browser.OpenURL(deviceResp.VerificationURIComplete); err != nil {
		view_common.RenderInfoMessage("Could not open browser automatically.")
		view_common.RenderInfoMessage("Please visit the URL above manually.")
	} else {
		view_common.RenderInfoMessage("Browser opened. Complete authentication there.")
	}

	fmt.Println()
	view_common.RenderInfoMessage("Waiting for authorization...")

	// Step 3: Poll for token
	token, err := pollForToken(apiURL, deviceResp)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Step 4: Save token to config
	fmt.Println()
	view_common.RenderInfoMessageBold("Authentication successful!")
	view_common.RenderInfoMessage(fmt.Sprintf("Token is connected to organization: %s", token.OrganizationName))

	// Determine profile name
	profileName := profileFlag
	if profileName == "" {
		profileName = "device-auth"
	}

	if err := saveTokenToConfig(apiURL, profileName, token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		configDir = "~/.config/daytona"
	}

	view_common.RenderInfoMessageBold(fmt.Sprintf("Token saved to %s/config.json in profile '%s'", configDir, profileName))

	return nil
}

func requestDeviceCode(apiURL string) (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/auth/device/code", apiURL)

	payload := map[string]string{
		"client_id": "daytona-cli",
		"scope":     "workspaces:read workspaces:write",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to request device code: %s", string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, err
	}

	return &deviceResp, nil
}

func pollForToken(apiURL string, deviceResp *DeviceCodeResponse) (*TokenResponse, error) {
	url := fmt.Sprintf("%s/auth/device/token", apiURL)
	interval := time.Duration(deviceResp.Interval) * time.Second
	expiresAt := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

	for {
		// Wait for the polling interval
		time.Sleep(interval)

		if time.Now().After(expiresAt) {
			return nil, fmt.Errorf("device code expired, please try again")
		}

		payload := map[string]string{
			"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
			"device_code": deviceResp.DeviceCode,
			"client_id":   "daytona-cli",
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Check if we got a successful token response
		if resp.StatusCode == http.StatusOK {
			// Try to parse as success response
			var tokenResp TokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				return nil, err
			}

			// Check if it's actually a success (has access_token)
			if tokenResp.AccessToken != "" {
				return &tokenResp, nil
			}

			// Otherwise it might be an error response
			var errResp TokenError
			if err := json.Unmarshal(body, &errResp); err != nil {
				return nil, err
			}

			switch errResp.Error {
			case "authorization_pending":
				// Keep waiting
				continue
			case "slow_down":
				// Increase polling interval
				interval += 5 * time.Second
				continue
			case "access_denied":
				return nil, fmt.Errorf("authorization denied by user")
			case "expired_token":
				return nil, fmt.Errorf("device code expired, please try again")
			case "":
				// No error, continue polling
				continue
			default:
				return nil, fmt.Errorf("unexpected error: %s", errResp.Error)
			}
		}

		// Parse error response
		var errResp TokenError
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("unexpected response: %s", string(body))
		}

		switch errResp.Error {
		case "authorization_pending":
			// Keep waiting
			continue
		case "slow_down":
			// Increase polling interval
			interval += 5 * time.Second
			continue
		case "access_denied":
			return nil, fmt.Errorf("authorization denied by user")
		case "expired_token":
			return nil, fmt.Errorf("device code expired, please try again")
		default:
			return nil, fmt.Errorf("unexpected error: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
	}
}

func saveTokenToConfig(apiURL string, profileName string, token *TokenResponse) error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	// Check if profile already exists
	existingProfile, _ := c.GetProfile(profileName)
	if existingProfile.Id != "" {
		// Update existing profile
		existingProfile.Api.Key = &token.AccessToken
		existingProfile.Api.Token = nil
		existingProfile.Api.Url = apiURL
		existingProfile.ActiveOrganizationId = &token.OrganizationID

		if err := c.EditProfile(existingProfile); err != nil {
			return err
		}
	} else {
		// Create new profile
		profile := config.Profile{
			Id:   profileName,
			Name: profileName,
			Api: config.ServerApi{
				Url: apiURL,
				Key: &token.AccessToken,
			},
			ActiveOrganizationId: &token.OrganizationID,
		}

		if err := c.AddProfile(profile); err != nil {
			return err
		}
	}

	// Get personal organization ID if needed (for browser-based OAuth this is different)
	// For API key auth, we already have the organization from the token response
	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return nil // Config saved successfully, this is just a follow-up action
	}

	// If we just set the profile as active, make sure we have the organization set
	if activeProfile.Id == profileName && activeProfile.ActiveOrganizationId == nil {
		activeProfile.ActiveOrganizationId = &token.OrganizationID
		if err := c.EditProfile(activeProfile); err != nil {
			return err
		}
	}

	// Try to get personal org if this is a new profile
	if activeProfile.Id == profileName {
		personalOrgId, err := common.GetPersonalOrganizationId(activeProfile)
		if err == nil && personalOrgId != "" {
			activeProfile.ActiveOrganizationId = &personalOrgId
			_ = c.EditProfile(activeProfile)
		}
	}

	return nil
}
