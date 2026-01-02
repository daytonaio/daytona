// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package vmpool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/runner-win/pkg/libvirt"
	log "github.com/sirupsen/logrus"
)

// RecoverExistingPoolVMs scans libvirt for existing paused VMs that can be added to the pool.
// This is called on startup to recover pool state from a previous run.
func (p *Pool) RecoverExistingPoolVMs(ctx context.Context) error {
	log.Info("Scanning for existing pool VMs to recover...")

	// List all domains
	domains, err := p.libvirt.DomainList(ctx, libvirt.DomainListOptions{All: true})
	if err != nil {
		return err
	}

	recovered := 0
	for _, domain := range domains {
		// Check if this looks like a pool VM (sndbx-* prefix from memory snapshot creation)
		// or a cold-boot pool VM (pool-vm-* prefix)
		isPoolVM := strings.HasPrefix(domain.Name, "sndbx-") || strings.HasPrefix(domain.Name, "pool-vm-")

		if !isPoolVM {
			continue
		}

		// Check if it's paused (ready to be claimed)
		if domain.State != libvirt.DomainStatePaused {
			log.Debugf("Skipping domain %s: not paused (state: %d)", domain.Name, domain.State)
			continue
		}

		// Check if it already has sandbox metadata (meaning it's claimed by a sandbox)
		metadata, err := p.libvirt.GetSandboxMetadata(ctx, domain.Name)
		if err == nil && metadata != "" && !strings.Contains(metadata, "pool-vm-") {
			// Has sandbox metadata that's not a pool-vm ID, so it's a claimed sandbox
			log.Debugf("Skipping domain %s: has sandbox metadata (claimed)", domain.Name)
			continue
		}

		// This is an unclaimed, paused pool VM - add it to the pool
		log.Infof("Recovering pool VM: %s (UUID: %s)", domain.Name, domain.UUID)

		// Get the IP address
		ip := p.libvirt.GetActualDomainIP(domain.Name)

		// For recovered VMs, the domain.Name IS the actual libvirt domain name
		// So both Name and DomainName should be set to the same value
		vm := &PoolVM{
			Name:       domain.Name, // Used as pool tracking key
			DomainName: domain.Name, // Actual libvirt domain name (same for recovered VMs)
			UUID:       domain.UUID,
			IP:         ip,
			MAC:        "", // Will be determined when claimed
			State:      PoolVMStateAvailable,
			CreatedAt:  time.Now(), // Approximate
		}

		p.mu.Lock()
		p.vms[domain.Name] = vm
		p.mu.Unlock()

		recovered++
		log.Infof("Recovered pool VM %s (domain: %s, IP: %s)", domain.Name, domain.Name, ip)
	}

	log.Infof("Pool recovery complete: recovered %d VMs", recovered)
	return nil
}

// UpdateNextIndex updates the next index counter based on existing pool VMs
// This prevents name collisions when creating new pool VMs
func (p *Pool) UpdateNextIndex() {
	p.mu.Lock()
	defer p.mu.Unlock()

	maxIndex := 0
	for name := range p.vms {
		// Try to extract index from pool-vm-XXXXXXX format
		if strings.HasPrefix(name, PoolVMNamePrefix) {
			var idx int
			_, err := fmt.Sscanf(name, PoolVMNamePrefix+"%d", &idx)
			if err == nil && idx > maxIndex {
				maxIndex = idx
			}
		}
	}

	// Set next index to be higher than any existing
	if maxIndex >= p.nextIndex {
		p.nextIndex = maxIndex + 1
		log.Infof("Updated pool VM next index to %d", p.nextIndex)
	}
}
