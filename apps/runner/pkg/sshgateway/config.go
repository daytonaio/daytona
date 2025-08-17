// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"fmt"
	"os"
	"strconv"
)

const (
	SSH_GATEWAY_PORT = 2220
)

// IsSSHGatewayEnabled checks if the SSH gateway should be enabled
func IsSSHGatewayEnabled() bool {
	return os.Getenv("SSH_GATEWAY_ENABLE") == "true"
}

// GetSSHGatewayPort returns the SSH gateway port
func GetSSHGatewayPort() int {
	if port := os.Getenv("SSH_GATEWAY_PORT"); port != "" {
		if parsedPort, err := strconv.Atoi(port); err == nil {
			return parsedPort
		}
	}
	return SSH_GATEWAY_PORT
}

// GetSSHPublicKey returns the SSH public key from configuration
func GetSSHPublicKey() (string, error) {
	publicKey := os.Getenv("SSH_PUBLIC_KEY")
	if publicKey == "" {
		return "", fmt.Errorf("SSH_PUBLIC_KEY environment variable not set")
	}
	return publicKey, nil
}

// GetSSHHostKey returns the SSH host key from configuration
func GetSSHHostKey() (string, error) {
	hostKey := os.Getenv("SSH_HOST_KEY")
	if hostKey == "" {
		return "", fmt.Errorf("SSH_HOST_KEY environment variable not set")
	}
	return hostKey, nil
}
