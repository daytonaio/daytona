// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"context"

	"github.com/daytonaio/mock-runner/pkg/toolbox"
	log "github.com/sirupsen/logrus"
)

// Service handles SSH gateway connections routing to the shared toolbox container
type Service struct {
	toolboxContainer *toolbox.ToolboxContainer
}

// NewService creates a new SSH gateway service
func NewService(toolboxContainer *toolbox.ToolboxContainer) *Service {
	return &Service{
		toolboxContainer: toolboxContainer,
	}
}

// Start starts the SSH gateway service
func (s *Service) Start(ctx context.Context) error {
	log.Info("Mock SSH Gateway: Starting (routes all SSH to toolbox container)")

	// In a real implementation, this would start an SSH server
	// that routes all connections to the toolbox container
	// For now, this is a placeholder that logs the intention

	toolboxIP := s.toolboxContainer.GetIP()
	log.Infof("Mock SSH Gateway: Would route SSH connections to toolbox container at %s", toolboxIP)

	// Block until context is done
	<-ctx.Done()
	return nil
}

// GetToolboxIP returns the IP of the toolbox container for SSH routing
func (s *Service) GetToolboxIP() string {
	return s.toolboxContainer.GetIP()
}



