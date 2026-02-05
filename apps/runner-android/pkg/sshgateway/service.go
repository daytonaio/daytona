// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	cvdClient *cuttlefish.Client
	port      int
}

func NewService(cvdClient *cuttlefish.Client) *Service {
	port := GetSSHGatewayPort()

	service := &Service{
		cvdClient: cvdClient,
		port:      port,
	}

	return service
}

// GetPort returns the port the SSH gateway is configured to use
func (s *Service) GetPort() int {
	return s.port
}

// Start starts the SSH gateway server
// Note: For Cuttlefish, SSH gateway is not the primary way to access Android devices.
// ADB is the recommended method for accessing Android instances.
func (s *Service) Start(ctx context.Context) error {
	log.Info("SSH Gateway: Not starting - use ADB for Cuttlefish Android devices")
	log.Info("SSH Gateway: For shell access, use: adb -s <serial> shell")

	// List available ADB serials
	sandboxes, err := s.cvdClient.ListWithInfo(ctx)
	if err == nil && len(sandboxes) > 0 {
		log.Info("SSH Gateway: Available ADB devices:")
		for _, s := range sandboxes {
			log.Infof("  - %s (ID: %s, State: %s)", s.ADBSerial, s.Id, s.State)
		}
	}

	// Keep the service running (do nothing)
	<-ctx.Done()
	return nil
}

// SandboxDetails contains information about a sandbox
type SandboxDetails struct {
	User      string `json:"user"`
	Hostname  string `json:"hostname"`
	ADBSerial string `json:"adbSerial"`
}

// getSandboxDetails gets sandbox information via Cuttlefish client
func (s *Service) getSandboxDetails(sandboxId string) (*SandboxDetails, error) {
	ctx := context.Background()

	sandboxInfo, err := s.cvdClient.GetSandboxInfo(ctx, sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox info for %s: %w", sandboxId, err)
	}

	return &SandboxDetails{
		User:      "shell", // ADB shell user
		ADBSerial: sandboxInfo.ADBSerial,
	}, nil
}
