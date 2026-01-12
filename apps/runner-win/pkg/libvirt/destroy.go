// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

func (l *LibVirt) Destroy(ctx context.Context, domainId string) error {
	domainMutex := l.getDomainMutex(domainId)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateDestroying)
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
			log.Warnf("Domain %s not found, considering it already destroyed", domainId)
			if l.statesCache != nil {
				l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateDestroyed)
			}
			return nil
		}
	}
	defer domain.Free()

	// Check if domain is running
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	// Force stop the domain if it's running
	if state != libvirt.DOMAIN_SHUTOFF {
		log.Infof("Force stopping domain %s before destroying", domainId)
		if err := domain.Destroy(); err != nil {
			log.Warnf("Failed to destroy running domain: %v", err)
			// Continue anyway to try undefining
		}
	}

	// Check and remove managed save image if it exists
	hasManagedSave, err := domain.HasManagedSaveImage(0)
	if err != nil {
		log.Warnf("Failed to check managed save for %s: %v", domainId, err)
	} else if hasManagedSave {
		log.Infof("Removing managed save image for domain %s", domainId)
		if err := domain.ManagedSaveRemove(0); err != nil {
			log.Warnf("Failed to remove managed save: %v", err)
			// Continue anyway - undefine with flag might still work
		}
	}

	// Undefine the domain with NVRAM and managed save removal flags
	log.Infof("Undefining domain %s with NVRAM and managed save flags", domainId)
	undefineFlags := libvirt.DOMAIN_UNDEFINE_NVRAM | libvirt.DOMAIN_UNDEFINE_MANAGED_SAVE
	if err := domain.UndefineFlags(undefineFlags); err != nil {
		// Try with just NVRAM flag
		log.Warnf("Failed to undefine with all flags, trying NVRAM only: %v", err)
		if err := domain.UndefineFlags(libvirt.DOMAIN_UNDEFINE_NVRAM); err != nil {
			// Try without any flags
			log.Warnf("Failed to undefine with NVRAM flag, trying without: %v", err)
			if err := domain.Undefine(); err != nil {
				return fmt.Errorf("failed to undefine domain: %w", err)
			}
		}
	}

	// Clean up disk and NVRAM files on remote host
	l.cleanupDomainFiles(context.Background(), domainId)

	// Clean up DHCP reservation to free the IP
	mac := GetReservedMAC(domainId)
	ip := GetReservedIP(domainId)
	if err := l.RemoveDHCPReservation(mac, ip); err != nil {
		log.Warnf("Failed to remove DHCP reservation for %s: %v", domainId, err)
		// Don't fail the destroy operation for DHCP cleanup failure
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, domainId, enums.SandboxStateDestroyed)
	}
	log.Infof("Domain %s destroyed successfully", domainId)

	return nil
}

func (l *LibVirt) RemoveDestroyed(ctx context.Context, domainId string) error {
	// For libvirt, destroying already undefines the domain
	// So this is essentially a no-op, but we keep it for interface compatibility
	log.Infof("RemoveDestroyed called for %s (no-op for libvirt)", domainId)
	return nil
}

// cleanupDomainFiles removes disk and NVRAM files for a destroyed domain
func (l *LibVirt) cleanupDomainFiles(ctx context.Context, domainId string) {
	isLocal := l.isLocalURI()
	host := ""
	if !isLocal {
		host = l.extractHostFromURI()
		if host == "" {
			log.Warnf("Could not extract host from URI for cleanup")
			return
		}
	}

	// Build paths for disk and NVRAM
	diskPath := filepath.Join(sandboxesBasePath, fmt.Sprintf("%s.qcow2", domainId))
	nvramPath := filepath.Join(nvramBasePath, fmt.Sprintf("%s_VARS.fd", domainId))

	// Remove disk file
	log.Infof("Removing disk file: %s", diskPath)
	var cmd *exec.Cmd
	if isLocal {
		cmd = exec.CommandContext(ctx, "rm", "-f", diskPath)
	} else {
		cmd = exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("rm -f %s", diskPath))
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warnf("Failed to remove disk %s: %v (output: %s)", diskPath, err, string(output))
	}

	// Remove NVRAM file
	log.Infof("Removing NVRAM file: %s", nvramPath)
	if isLocal {
		cmd = exec.CommandContext(ctx, "rm", "-f", nvramPath)
	} else {
		cmd = exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("rm -f %s", nvramPath))
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Warnf("Failed to remove NVRAM %s: %v (output: %s)", nvramPath, err, string(output))
	}

	log.Infof("Cleanup completed for domain %s", domainId)
}
