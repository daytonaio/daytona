/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type SSHAccessValidation struct {
	Valid        bool   `json:"valid"`
	SandboxID    string `json:"sandboxId"`
	RunnerID     string `json:"runnerId,omitempty"`
	RunnerDomain string `json:"runnerDomain,omitempty"`
}

type RunnerSSHKeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type SandboxState struct {
	State string `json:"state"`
}

func NewAPIClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) ValidateSSHAccess(token string) (*SSHAccessValidation, error) {
	url := fmt.Sprintf("%s/sandbox/ssh-access/validate?token=%s", c.baseURL, token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var validation SSHAccessValidation
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &validation, nil
}

func (c *APIClient) GetRunnerSSHKeyPair(runnerID string) (*RunnerSSHKeyPair, error) {
	url := fmt.Sprintf("%s/runners/%s/ssh-keypair", c.baseURL, runnerID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var keypair RunnerSSHKeyPair
	if err := json.NewDecoder(resp.Body).Decode(&keypair); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &keypair, nil
}

func (c *APIClient) GetSandboxState(sandboxID string) (*SandboxState, error) {
	url := fmt.Sprintf("%s/sandbox/%s", c.baseURL, sandboxID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sandboxState SandboxState
	if err := json.NewDecoder(resp.Body).Decode(&sandboxState); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sandboxState, nil
}
