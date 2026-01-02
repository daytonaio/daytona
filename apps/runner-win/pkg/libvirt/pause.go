// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"

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
