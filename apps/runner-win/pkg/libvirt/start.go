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

func (l *LibVirt) Start(ctx context.Context, domainId string, metadata map[string]string) (string, error) {
	domainMutex := l.getDomainMutex(domainId)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStarting)
	}

	conn, err := l.getConnection()
	if err != nil {
		return "", fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain by UUID first, then by name
	domain, err := conn.LookupDomainByUUIDString(domainId)
	if err != nil {
		// Try by name
		domain, err = conn.LookupDomainByName(domainId)
		if err != nil {
			return "", fmt.Errorf("domain not found: %w", err)
		}
	}
	defer domain.Free()

	// Check if already running
	state, _, err := domain.GetState()
	if err != nil {
		return "", fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		log.Infof("Domain %s is already running", domainId)
		if l.statesCache != nil {
			l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStarted)
		}
		return l.getDaemonVersion(ctx, domain)
	}

	// Start the domain
	log.Infof("Starting domain %s", domainId)
	if err := domain.Create(); err != nil {
		return "", fmt.Errorf("failed to start domain: %w", err)
	}

	// Wait for domain to be running
	if err := l.waitForDomainRunning(ctx, domain); err != nil {
		return "", fmt.Errorf("domain failed to start: %w", err)
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStarted)
	}
	log.Infof("Domain %s started successfully", domainId)

	return l.getDaemonVersion(ctx, domain)
}

func (l *LibVirt) waitForDomainRunning(ctx context.Context, domain *libvirt.Domain) error {
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

		if state == libvirt.DOMAIN_RUNNING {
			return nil
		}

		if state == libvirt.DOMAIN_CRASHED || state == libvirt.DOMAIN_SHUTOFF {
			return fmt.Errorf("domain entered unexpected state: %d", state)
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for domain to start after %v", timeout)
}

func (l *LibVirt) getDaemonVersion(ctx context.Context, domain *libvirt.Domain) (string, error) {
	// For now, return a placeholder version
	// In the future, we might want to query the VM's daemon via network
	name, err := domain.GetName()
	if err != nil {
		return "unknown", nil
	}
	return fmt.Sprintf("libvirt-domain-%s", name), nil
}
