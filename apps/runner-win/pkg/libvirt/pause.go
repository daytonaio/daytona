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

// Pause suspends a running domain (freezes CPU execution but keeps memory)
func (l *LibVirt) Pause(ctx context.Context, domainId string) error {
	domainMutex := l.getDomainMutex(domainId)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain (handles both cold boot and memory snapshot naming)
	domain, err := l.LookupDomainBySandboxId(conn, domainId)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Check current state
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_PAUSED {
		log.Infof("Domain %s is already paused", domainId)
		return nil
	}

	if state != libvirt.DOMAIN_RUNNING {
		return fmt.Errorf("domain %s is not running (state: %d), cannot pause", domainId, state)
	}

	// Suspend the domain
	log.Infof("Pausing domain %s", domainId)
	if err := domain.Suspend(); err != nil {
		return fmt.Errorf("failed to pause domain: %w", err)
	}

	log.Infof("Domain %s paused successfully", domainId)
	return nil
}

// Resume resumes a paused domain
func (l *LibVirt) Resume(ctx context.Context, domainId string) error {
	domainMutex := l.getDomainMutex(domainId)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain (handles both cold boot and memory snapshot naming)
	domain, err := l.LookupDomainBySandboxId(conn, domainId)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Check current state
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		log.Infof("Domain %s is already running", domainId)
		return nil
	}

	if state != libvirt.DOMAIN_PAUSED {
		return fmt.Errorf("domain %s is not paused (state: %d), cannot resume", domainId, state)
	}

	// Resume the domain
	log.Infof("Resuming domain %s", domainId)
	if err := domain.Resume(); err != nil {
		return fmt.Errorf("failed to resume domain: %w", err)
	}

	log.Infof("Domain %s resumed successfully", domainId)
	return nil
}

// SetSandboxMetadata sets sandbox ID metadata on a domain for debugging purposes
// domainName is the actual libvirt domain name (e.g., "sndbx-3d4fcb3be")
// sandboxId is the sandbox ID to store in the metadata
func (l *LibVirt) SetSandboxMetadata(ctx context.Context, domainName string, sandboxId string) error {
	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Look up domain directly by name - we can't use LookupDomainBySandboxId here
	// because we're SETTING the metadata, it doesn't exist yet
	domain, err := conn.LookupDomainByName(domainName)
	if err != nil {
		return fmt.Errorf("domain not found by name %s: %w", domainName, err)
	}
	defer domain.Free()

	log.Debugf("SetSandboxMetadata: Found domain %s, will set sandboxId=%s", domainName, sandboxId)

	// Check if domain is transient (created by virsh restore) and make it persistent
	// Transient domains don't support DOMAIN_AFFECT_CONFIG, so we need to define them first
	isPersistent, err := domain.IsPersistent()
	if err != nil {
		log.Warnf("Failed to check if domain %s is persistent: %v", domainName, err)
	} else if !isPersistent {
		log.Infof("Domain %s is transient, making it persistent to support metadata", domainName)
		// Get the current XML and define the domain to make it persistent
		xmlDesc, err := domain.GetXMLDesc(0)
		if err != nil {
			log.Warnf("Failed to get domain %s XML for making persistent: %v", domainName, err)
		} else {
			_, err = conn.DomainDefineXML(xmlDesc)
			if err != nil {
				log.Warnf("Failed to make domain %s persistent: %v", domainName, err)
			} else {
				log.Infof("Domain %s is now persistent", domainName)
			}
		}
	}

	// Set metadata with sandbox ID
	metadata := fmt.Sprintf(`<sandbox xmlns="http://daytona.io/sandbox"><id>%s</id></sandbox>`, sandboxId)

	// Set metadata on BOTH live domain AND config to persist across restarts
	// Try live first (for running domains)
	liveErr := domain.SetMetadata(
		libvirt.DOMAIN_METADATA_ELEMENT,
		metadata,
		"sandbox",
		"http://daytona.io/sandbox",
		libvirt.DOMAIN_AFFECT_LIVE,
	)

	// Always also set config to persist metadata across VM restarts
	configErr := domain.SetMetadata(
		libvirt.DOMAIN_METADATA_ELEMENT,
		metadata,
		"sandbox",
		"http://daytona.io/sandbox",
		libvirt.DOMAIN_AFFECT_CONFIG,
	)

	// If both fail, return error
	if liveErr != nil && configErr != nil {
		return fmt.Errorf("failed to set sandbox metadata (live: %v, config: %v)", liveErr, configErr)
	}

	log.Infof("Set sandbox metadata on domain %s: sandboxId=%s (live=%v, config=%v)",
		domainName, sandboxId, liveErr == nil, configErr == nil)
	return nil
}

// GetSandboxMetadata retrieves sandbox ID metadata from a domain
func (l *LibVirt) GetSandboxMetadata(ctx context.Context, domainId string) (string, error) {
	conn, err := l.getConnection()
	if err != nil {
		return "", fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain (handles both cold boot and memory snapshot naming)
	domain, err := l.LookupDomainBySandboxId(conn, domainId)
	if err != nil {
		return "", fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Try to get metadata from live domain first
	metadata, err := domain.GetMetadata(
		libvirt.DOMAIN_METADATA_ELEMENT,
		"http://daytona.io/sandbox",
		libvirt.DOMAIN_AFFECT_LIVE,
	)
	if err != nil {
		// Try config if live fails
		metadata, err = domain.GetMetadata(
			libvirt.DOMAIN_METADATA_ELEMENT,
			"http://daytona.io/sandbox",
			libvirt.DOMAIN_AFFECT_CONFIG,
		)
		if err != nil {
			return "", fmt.Errorf("failed to get sandbox metadata: %w", err)
		}
	}

	return metadata, nil
}

// SuspendToDisk saves the domain's memory state to disk and stops the domain.
// This frees up system RAM while preserving the exact VM state.
// The domain transitions to SHUTOFF state with a managed save image.
func (l *LibVirt) SuspendToDisk(ctx context.Context, domainId string) error {
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

	// Try to find domain (handles both cold boot and memory snapshot naming)
	domain, err := l.LookupDomainBySandboxId(conn, domainId)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Check current state
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	// If already shut off, check if it has a managed save image
	if state == libvirt.DOMAIN_SHUTOFF {
		hasManagedSave, err := domain.HasManagedSaveImage(0)
		if err == nil && hasManagedSave {
			log.Infof("Domain %s is already suspended to disk", domainId)
			if l.statesCache != nil {
				l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStopped)
			}
			return nil
		}
		log.Infof("Domain %s is already stopped (no managed save image)", domainId)
		if l.statesCache != nil {
			l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStopped)
		}
		return nil
	}

	if state != libvirt.DOMAIN_RUNNING && state != libvirt.DOMAIN_PAUSED {
		return fmt.Errorf("domain %s is not running or paused (state: %d), cannot suspend to disk", domainId, state)
	}

	// Save domain state to disk using ManagedSave
	// This saves memory to disk and stops the domain, freeing system RAM
	log.Infof("Suspending domain %s to disk (saving memory state)", domainId)
	if err := domain.ManagedSave(0); err != nil {
		return fmt.Errorf("failed to suspend domain to disk: %w", err)
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStopped)
	}
	log.Infof("Domain %s suspended to disk successfully", domainId)
	return nil
}

// ResumeFromDisk restores a domain from its managed save image on disk.
// If the domain has a saved state, it will be restored exactly as it was.
// Returns the daemon version after the domain is running.
func (l *LibVirt) ResumeFromDisk(ctx context.Context, domainId string, metadata map[string]string) (string, error) {
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

	// Try to find domain (handles both cold boot and memory snapshot naming)
	domain, err := l.LookupDomainBySandboxId(conn, domainId)
	if err != nil {
		return "", fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Get domain name for daemon readiness check
	domainName, err := domain.GetName()
	if err != nil {
		return "", fmt.Errorf("failed to get domain name: %w", err)
	}

	// Check current state
	state, _, err := domain.GetState()
	if err != nil {
		return "", fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		log.Infof("Domain %s is already running", domainId)
		// Even if VM is running, wait for daemon to be ready before returning
		if err := l.waitForDaemonReady(ctx, domainName, ""); err != nil {
			log.Warnf("Daemon readiness check failed for already running domain %s: %v", domainId, err)
			// Don't fail - the VM is running, daemon might just be slow
		}
		if l.statesCache != nil {
			l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStarted)
		}
		return l.getDaemonVersion(ctx, domainName)
	}

	// Check if domain has a managed save image
	hasManagedSave, err := domain.HasManagedSaveImage(0)
	if err != nil {
		log.Warnf("Failed to check managed save image for domain %s: %v", domainId, err)
	}

	if hasManagedSave {
		log.Infof("Resuming domain %s from disk (restoring saved memory state)", domainId)
	} else {
		log.Infof("Starting domain %s (no saved state found, cold boot)", domainId)
	}

	// Create (start) the domain - libvirt automatically restores from managed save if available
	if err := domain.Create(); err != nil {
		return "", fmt.Errorf("failed to resume domain from disk: %w", err)
	}

	// Wait for domain to be running at hypervisor level
	if err := l.waitForDomainRunningWithDomain(ctx, domain); err != nil {
		return "", fmt.Errorf("domain failed to start: %w", err)
	}

	log.Infof("Domain %s is running, waiting for daemon to be ready...", domainId)

	// Wait for daemon inside VM to be ready to accept connections
	if err := l.waitForDaemonReady(ctx, domainName, ""); err != nil {
		return "", fmt.Errorf("daemon failed to become ready: %w", err)
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateStarted)
	}
	log.Infof("Domain %s resumed from disk successfully and daemon is ready", domainId)

	return l.getDaemonVersion(ctx, domainName)
}

// waitForDomainRunningWithDomain waits for a domain to reach running state
func (l *LibVirt) waitForDomainRunningWithDomain(ctx context.Context, domain *libvirt.Domain) error {
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

// WaitForDaemonReady is an exported version that waits for the daemon API to be accessible
// Used by the VM pool watcher to ensure daemon is ready before pausing
func (l *LibVirt) WaitForDaemonReady(ctx context.Context, domainId string) error {
	// Get the IP for this domain
	ip := l.getActualDomainIP(domainId)
	return l.waitForDaemonReady(ctx, domainId, ip)
}

// GetActualDomainIP returns the actual IP address of a domain from DHCP lease
func (l *LibVirt) GetActualDomainIP(domainId string) string {
	return l.getActualDomainIP(domainId)
}
