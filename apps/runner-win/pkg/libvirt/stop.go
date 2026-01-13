// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

func (l *LibVirt) Stop(ctx context.Context, domainId string) error {
	domainMutex := l.getDomainMutex(domainId)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStopping)
	}

	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain by UUID first, then by name
	domain, err := conn.LookupDomainByUUIDString(domainId)
	if err != nil {
		// Try by name
		domain, err = conn.LookupDomainByName(domainId)
		if err != nil {
			return fmt.Errorf("domain not found: %w", err)
		}
	}
	defer domain.Free()

	// Check current state
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_SHUTOFF {
		log.Infof("Domain %s is already stopped", domainId)
		if l.statesCache != nil {
			l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStopped)
		}
		return nil
	}

	// Try graceful shutdown via Windows shutdown command first
	// This avoids the "Display Shutdown Event Tracker" dialog
	log.Infof("Attempting graceful shutdown of domain %s via guest command", domainId)
	if err := l.ShutdownGuest(ctx, domainId); err != nil {
		log.Warnf("Guest shutdown command failed: %v, trying libvirt shutdown", err)
		// Fall back to libvirt shutdown
		if err := domain.Shutdown(); err != nil {
			log.Warnf("Libvirt shutdown also failed: %v, attempting force destroy", err)
			// If graceful shutdown fails, force destroy
			if err := domain.Destroy(); err != nil {
				return fmt.Errorf("failed to force stop domain: %w", err)
			}
		}
	}

	// Wait for domain to be stopped after graceful shutdown
	if err := l.waitForDomainStopped(ctx, domain); err != nil {
		// Graceful shutdown timed out, force destroy
		log.Warnf("Graceful shutdown timed out for domain %s: %v, forcing destroy", domainId, err)
		if err := domain.Destroy(); err != nil {
			// Domain might already be stopped or in transition, check state
			state, _, stateErr := domain.GetState()
			if stateErr != nil || state != libvirt.DOMAIN_SHUTOFF {
				return fmt.Errorf("failed to force stop domain after timeout: %w", err)
			}
			// Domain is already stopped, continue
			log.Infof("Domain %s is already stopped", domainId)
			return nil
		}
		// Wait again after force destroy
		if err := l.waitForDomainStopped(ctx, domain); err != nil {
			return fmt.Errorf("domain failed to stop after force destroy: %w", err)
		}
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStopped)
	}
	log.Infof("Domain %s stopped successfully", domainId)

	return nil
}

func (l *LibVirt) waitForDomainStopped(ctx context.Context, domain *libvirt.Domain) error {
	timeout := time.Duration(l.sandboxStartTimeoutSec) * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		state, _, err := domain.GetState()
		if err != nil {
			return fmt.Errorf("failed to get domain state: %w", err)
		}

		if state == libvirt.DOMAIN_SHUTOFF {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for domain to stop after %v", timeout)
}
