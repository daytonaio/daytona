// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/pkg/browser"
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
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	ExpiresIn      int    `json:"expires_in"`
	Scope          string `json:"scope"`
	OrganizationID string `json:"organization_id"`
	User           struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func loginWithDeviceFlow(ctx context.Context) (*config.Token, error) {
	apiURL := config.GetDaytonaApiUrl()

	// Step 1: Request device code
	view_common.RenderInfoMessageBold("Requesting device authorization...")

	deviceResp, err := requestDeviceCode(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Display code and open browser
	fmt.Println()
	view_common.RenderInfoMessage("Please visit this URL in your browser:\n")
	fmt.Printf("\n    %s\n\n", deviceResp.VerificationURIComplete)
	view_common.RenderInfoMessage(fmt.Sprintf("Or go to %s and enter code:\n", deviceResp.VerificationURI))
	fmt.Printf("\n    %s\n\n", deviceResp.UserCode)

	// Try to open browser automatically
	view_common.RenderInfoMessageBold("Opening the browser for authentication ...")
	if err := browser.OpenURL(deviceResp.VerificationURIComplete); err != nil {
		fmt.Println("Was not able to launch web browser")
		fmt.Printf("Please go to this URL manually and complete the flow:\n\n%s\n\n",
			deviceResp.VerificationURIComplete)
	}

	// Step 3: Poll for token
	tokenResp, err := pollForToken(ctx, apiURL, deviceResp)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	view_common.RenderInfoMessageBold("Web authentication finished successfully!")
	fmt.Printf("Token is connected to the %s organization.\n", tokenResp.OrganizationID)

	return &config.Token{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

func requestDeviceCode(apiURL string) (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/device/code", apiURL)

	payload := map[string]string{
		"client_id": "daytona-cli",
		"scope":     "sandboxes:read sandboxes:write",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to request device code: %s - %s", resp.Status, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, err
	}

	return &deviceResp, nil
}

func pollForToken(ctx context.Context, apiURL string, deviceResp *DeviceCodeResponse) (*TokenResponse, error) {
	url := fmt.Sprintf("%s/device/token", apiURL)
	interval := time.Duration(deviceResp.Interval) * time.Second
	expiresAt := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-ticker.C:
			payload := map[string]string{
				"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
				"device_code": deviceResp.DeviceCode,
				"client_id":   "daytona-cli",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}

			req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}

			// Check if we got a token
			if resp.StatusCode == http.StatusOK {
				var tokenResp TokenResponse
				if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
					resp.Body.Close()
					return nil, err
				}
				resp.Body.Close()
				return &tokenResp, nil
			}

			// Check for errors
			var errResp TokenError
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()

			switch errResp.Error {
			case "authorization_pending":
				// Keep waiting
				continue
			case "slow_down":
				// Increase polling interval
				interval += 5 * time.Second
				ticker.Reset(interval)
				continue
			case "access_denied":
				return nil, fmt.Errorf("authorization denied by user")
			case "expired_token":
				return nil, fmt.Errorf("device code expired")
			default:
				if errResp.ErrorDescription != "" {
					return nil, fmt.Errorf("%s: %s", errResp.Error, errResp.ErrorDescription)
				}
				return nil, fmt.Errorf("unexpected error: %s", errResp.Error)
			}

		case <-time.After(time.Until(expiresAt)):
			return nil, fmt.Errorf("authentication timeout")
		}
	}
}
