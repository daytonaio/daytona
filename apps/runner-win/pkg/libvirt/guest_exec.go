// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// GuestExecRequest represents a command execution request to the daemon
type GuestExecRequest struct {
	Command string  `json:"command"`
	Timeout *uint32 `json:"timeout,omitempty"`
	Cwd     *string `json:"cwd,omitempty"`
}

// GuestExecResponse represents the response from a command execution
type GuestExecResponse struct {
	ExitCode int    `json:"exitCode"`
	Result   string `json:"result"`
}

// ExecuteGuestCommand executes a command inside the guest VM via the daemon's API
// This is useful for executing Windows commands like "shutdown /s /t 0" inside the VM
func (l *LibVirt) ExecuteGuestCommand(ctx context.Context, domainId, command string, timeoutSec uint32) (*GuestExecResponse, error) {
	// Get the VM's IP address
	actualIP := l.getActualDomainIP(domainId)
	if actualIP == "" {
		return nil, fmt.Errorf("could not get IP for domain %s", domainId)
	}

	log.Infof("Executing command in guest %s (IP: %s): %s", domainId, actualIP, command)

	// Prepare request body
	reqBody := GuestExecRequest{
		Command: command,
	}
	if timeoutSec > 0 {
		reqBody.Timeout = &timeoutSec
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the URL to the daemon's execute endpoint
	daemonURL := fmt.Sprintf("http://%s:2280/process/execute", actualIP)

	// Create HTTP client with SSH tunnel if remote
	var client *http.Client
	if IsRemoteURI(l.libvirtURI) {
		sshHost := l.extractHostFromURI()
		transport := GetSSHTunnelTransport(sshHost)
		client = &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeoutSec+30) * time.Second, // Extra time for network overhead
		}
	} else {
		client = &http.Client{
			Timeout: time.Duration(timeoutSec+30) * time.Second,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, daemonURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}

	var execResp GuestExecResponse
	if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Infof("Command executed in guest %s: exit_code=%d, output=%s", domainId, execResp.ExitCode, execResp.Result)

	return &execResp, nil
}

// ShutdownGuest sends the Windows shutdown command to the guest VM
// This avoids the "Display Shutdown Event Tracker" dialog by using "shutdown /s /t 0"
func (l *LibVirt) ShutdownGuest(ctx context.Context, domainId string) error {
	// The shutdown command:
	// /s = Shutdown the computer
	// /t 0 = Time-out period before shutdown (0 = immediate)
	// This avoids the "Display Shutdown Event Tracker" dialog
	_, err := l.ExecuteGuestCommand(ctx, domainId, "shutdown /s /t 0", 30)
	if err != nil {
		// The command might "fail" because the VM shuts down before responding
		// This is actually expected behavior
		log.Warnf("ShutdownGuest command returned error (expected if VM is shutting down): %v", err)
	}
	return nil
}
