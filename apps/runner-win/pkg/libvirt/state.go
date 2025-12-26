// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
)

func (l *LibVirt) DeduceSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	conn, err := l.getConnection()
	if err != nil {
		return enums.SandboxStateUnknown, fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain by UUID first, then by name
	domain, err := conn.LookupDomainByUUIDString(sandboxId)
	if err != nil {
		// Try by name
		domain, err = conn.LookupDomainByName(sandboxId)
		if err != nil {
			log.Debugf("Domain %s not found: %v", sandboxId, err)
			return enums.SandboxStateDestroyed, nil
		}
	}
	defer domain.Free()

	state, _, err := domain.GetState()
	if err != nil {
		return enums.SandboxStateUnknown, fmt.Errorf("failed to get domain state: %w", err)
	}

	return mapDomainStateToSandboxState(DomainState(state)), nil
}

// mapDomainStateToSandboxState maps libvirt domain states to Daytona sandbox states
func mapDomainStateToSandboxState(state DomainState) enums.SandboxState {
	switch state {
	case DomainStateRunning:
		return enums.SandboxStateStarted
	case DomainStateShutoff:
		return enums.SandboxStateStopped
	case DomainStateShutdown:
		return enums.SandboxStateStopping
	case DomainStatePaused:
		return enums.SandboxStateStopped
	case DomainStateBlocked:
		return enums.SandboxStateStarted // Blocked but running
	case DomainStateCrashed:
		return enums.SandboxStateDestroyed
	case DomainStatePMSuspended:
		return enums.SandboxStateStopped
	case DomainStateNoState:
		return enums.SandboxStateUnknown
	default:
		return enums.SandboxStateUnknown
	}
}

// GetDomainState returns the current state of a domain
func (l *LibVirt) GetDomainState(ctx context.Context, domainName string) (DomainState, error) {
	conn, err := l.getConnection()
	if err != nil {
		return DomainStateNoState, fmt.Errorf("failed to get connection: %w", err)
	}

	domain, err := conn.LookupDomainByName(domainName)
	if err != nil {
		return DomainStateNoState, fmt.Errorf("failed to lookup domain: %w", err)
	}
	defer domain.Free()

	state, _, err := domain.GetState()
	if err != nil {
		return DomainStateNoState, fmt.Errorf("failed to get domain state: %w", err)
	}

	return DomainState(state), nil
}
