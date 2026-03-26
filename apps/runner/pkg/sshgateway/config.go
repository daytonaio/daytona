// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/crypto/ssh"
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
// Get ssh public key of ssh gateway - external ssh gateway, not runner ssh gateway
// Should be base64 encoded
func GetSSHPublicKey() (string, error) {
	publicKey := os.Getenv("SSH_PUBLIC_KEY")
	if publicKey == "" {
		return "", fmt.Errorf("SSH_PUBLIC_KEY environment variable not set")
	}

	decodedKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode SSH_PUBLIC_KEY: %w", err)
	}

	return string(decodedKey), nil
}

// GetSSHHostKeyPath returns the SSH host key path from configuration
func GetSSHHostKeyPath() string {
	if path := os.Getenv("SSH_HOST_KEY_PATH"); path != "" {
		return path
	}
	return "/root/.ssh/id_rsa"
}

// GetSSHHostKey returns the SSH host key from configuration
func GetSSHHostKey() (ssh.Signer, error) {
	hostKeyPath := GetSSHHostKeyPath()

	// Check if host key file exists
	if _, err := os.Stat(hostKeyPath); os.IsNotExist(err) {
		// Generate new host key if file doesn't exist
		if err := generateAndSaveHostKey(hostKeyPath); err != nil {
			return nil, fmt.Errorf("failed to generate host key: %w", err)
		}
	}

	// Read the host key file
	hostKeyBytes, err := os.ReadFile(hostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key file: %w", err)
	}

	// Parse the private key to get the public key
	signer, err := ssh.ParsePrivateKey(hostKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host key: %w", err)
	}

	return signer, nil
}

// generateAndSaveHostKey generates a new RSA host key and saves it to the file
func generateAndSaveHostKey(hostKeyPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(hostKeyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate a new RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Save private key to file
	privateKeyFile, err := os.OpenFile(hostKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create host key file: %w", err)
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	return nil
}
