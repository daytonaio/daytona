// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

	// Get domain name for daemon readiness check
	domainName, err := domain.GetName()
	if err != nil {
		return "", fmt.Errorf("failed to get domain name: %w", err)
	}

	// Check if already running
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

	// Start the domain
	log.Infof("Starting domain %s", domainId)
	if err := domain.Create(); err != nil {
		return "", fmt.Errorf("failed to start domain: %w", err)
	}

	// Wait for domain to be running at hypervisor level
	if err := l.waitForDomainRunning(ctx, domain); err != nil {
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
	log.Infof("Domain %s started successfully and daemon is ready", domainId)

	return l.getDaemonVersion(ctx, domainName)
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

func (l *LibVirt) getDaemonVersion(ctx context.Context, domainName string) (string, error) {
	// Get the actual IP address from DHCP lease
	actualIP := l.getActualDomainIP(domainName)
	if actualIP == "" {
		log.Warnf("Could not get IP for domain %s, returning unknown version", domainName)
		return "unknown", nil
	}

	// Create HTTP client with SSH tunnel if remote
	var client *http.Client
	if IsRemoteURI(l.libvirtURI) {
		sshHost := l.extractHostFromURI()
		transport := GetSSHTunnelTransport(sshHost)
		client = &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		}
	} else {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	// Query the daemon's version endpoint
	daemonURL := fmt.Sprintf("http://%s:2280/version", actualIP)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, daemonURL, nil)
	if err != nil {
		log.Warnf("Failed to create version request for domain %s: %v", domainName, err)
		return "unknown", nil
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Warnf("Failed to get daemon version for domain %s: %v", domainName, err)
		return "unknown", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("Daemon version endpoint returned status %d for domain %s", resp.StatusCode, domainName)
		return "unknown", nil
	}

	// Parse version response
	var versionResp struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
		log.Warnf("Failed to decode version response for domain %s: %v", domainName, err)
		return "unknown", nil
	}

	if versionResp.Version == "" {
		return "unknown", nil
	}

	return versionResp.Version, nil
}
