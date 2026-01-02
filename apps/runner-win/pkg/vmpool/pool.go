// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package vmpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/runner-win/pkg/libvirt"
	log "github.com/sirupsen/logrus"
)

// Pool manages a pool of pre-created, paused VMs for fast sandbox creation
type Pool struct {
	mu      sync.RWMutex
	vms     map[string]*PoolVM // keyed by VM name
	config  PoolConfig
	libvirt *libvirt.LibVirt

	// Index counter for generating unique VM names
	nextIndex int
}

// NewPool creates a new VM pool
func NewPool(config PoolConfig, libvirtClient *libvirt.LibVirt) *Pool {
	return &Pool{
		vms:       make(map[string]*PoolVM),
		config:    config,
		libvirt:   libvirtClient,
		nextIndex: 1,
	}
}

// Claim attempts to claim an available VM from the pool for a sandbox
// Returns the claimed VM or nil if no VMs are available
func (p *Pool) Claim(ctx context.Context, sandboxID string) (*PoolVM, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find an available VM
	for _, vm := range p.vms {
		if vm.State == PoolVMStateAvailable {
			log.Infof("Claiming pool VM %s (domain: %s) for sandbox %s", vm.Name, vm.DomainName, sandboxID)

			// Resume the VM using the actual libvirt domain name
			if err := p.libvirt.Resume(ctx, vm.DomainName); err != nil {
				log.Errorf("Failed to resume pool VM %s (domain: %s): %v", vm.Name, vm.DomainName, err)
				// Mark as creating so watcher will try again or clean up
				vm.State = PoolVMStateCreating
				continue
			}

			// Set sandbox metadata for debugging using actual domain name
			if err := p.libvirt.SetSandboxMetadata(ctx, vm.DomainName, sandboxID); err != nil {
				log.Warnf("Failed to set sandbox metadata on VM %s (domain: %s): %v", vm.Name, vm.DomainName, err)
				// Continue anyway, metadata is just for debugging
			}

			// Update VM state
			vm.State = PoolVMStateClaimed
			vm.SandboxID = sandboxID
			vm.ClaimedAt = time.Now()

			log.Infof("Successfully claimed pool VM %s (domain: %s) for sandbox %s", vm.Name, vm.DomainName, sandboxID)
			return vm, nil
		}
	}

	log.Debugf("No available VMs in pool for sandbox %s", sandboxID)
	return nil, nil
}

// Add adds a VM to the pool as available
func (p *Pool) Add(vm *PoolVM) {
	p.mu.Lock()
	defer p.mu.Unlock()

	vm.State = PoolVMStateAvailable
	vm.CreatedAt = time.Now()
	p.vms[vm.Name] = vm
	log.Infof("Added VM %s to pool (available: %d)", vm.Name, p.countAvailable())
}

// Remove removes a VM from the pool
func (p *Pool) Remove(vmName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.vms, vmName)
	log.Infof("Removed VM %s from pool", vmName)
}

// Release releases a claimed VM back to the pool (for cleanup/destroy)
// This is called when a sandbox is destroyed to remove the VM from tracking
func (p *Pool) Release(sandboxID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, vm := range p.vms {
		if vm.SandboxID == sandboxID {
			delete(p.vms, name)
			log.Infof("Released VM %s from pool (sandbox %s destroyed)", name, sandboxID)
			return
		}
	}
}

// GetBySandboxID returns the VM claimed by a specific sandbox
func (p *Pool) GetBySandboxID(sandboxID string) *PoolVM {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, vm := range p.vms {
		if vm.SandboxID == sandboxID {
			return vm
		}
	}
	return nil
}

// GetByName returns a VM by its name
func (p *Pool) GetByName(name string) *PoolVM {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.vms[name]
}

// Stats returns current pool statistics
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		TargetSize: p.config.Size,
	}

	for _, vm := range p.vms {
		switch vm.State {
		case PoolVMStateAvailable:
			stats.Available++
		case PoolVMStateClaimed:
			stats.Claimed++
		case PoolVMStateCreating:
			stats.Creating++
		}
	}

	return stats
}

// NeedsReplenishment returns true if the pool needs more VMs
func (p *Pool) NeedsReplenishment() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	available := p.countAvailable()
	creating := p.countCreating()

	// Need more VMs if available + creating < target
	return (available + creating) < p.config.Size
}

// GetConfig returns the pool configuration
func (p *Pool) GetConfig() PoolConfig {
	return p.config
}

// GenerateVMName generates a unique pool VM name
func (p *Pool) GenerateVMName() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := p.generateVMNameLocked()
	p.nextIndex++
	return name
}

// generateVMNameLocked generates a VM name (must hold lock)
func (p *Pool) generateVMNameLocked() string {
	// Format: pool-vm-XXX where XXX is a zero-padded number
	// Total length: 15 chars to match golden template length
	// "pool-vm-" = 8 chars, leaving 7 chars for the number
	return fmt.Sprintf("%s%07d", PoolVMNamePrefix, p.nextIndex)
}

// MarkCreating marks a VM name as being created (to avoid duplicate creation)
func (p *Pool) MarkCreating(vmName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.vms[vmName] = &PoolVM{
		Name:      vmName,
		State:     PoolVMStateCreating,
		CreatedAt: time.Now(),
	}
}

// FinishCreating updates a VM that was being created to available
func (p *Pool) FinishCreating(vmName, domainName, uuid, ip, mac string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if vm, ok := p.vms[vmName]; ok {
		vm.DomainName = domainName
		vm.UUID = uuid
		vm.IP = ip
		vm.MAC = mac
		vm.State = PoolVMStateAvailable
		log.Infof("Pool VM %s (domain: %s) is now available (IP: %s)", vmName, domainName, ip)
	}
}

// CancelCreating removes a VM that failed during creation
func (p *Pool) CancelCreating(vmName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if vm, ok := p.vms[vmName]; ok && vm.State == PoolVMStateCreating {
		delete(p.vms, vmName)
		log.Infof("Cancelled creation of pool VM %s", vmName)
	}
}

// countAvailable counts available VMs (must hold at least read lock)
func (p *Pool) countAvailable() int {
	count := 0
	for _, vm := range p.vms {
		if vm.State == PoolVMStateAvailable {
			count++
		}
	}
	return count
}

// countCreating counts VMs being created (must hold at least read lock)
func (p *Pool) countCreating() int {
	count := 0
	for _, vm := range p.vms {
		if vm.State == PoolVMStateCreating {
			count++
		}
	}
	return count
}

// ListAvailable returns a list of available VM names
func (p *Pool) ListAvailable() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var names []string
	for _, vm := range p.vms {
		if vm.State == PoolVMStateAvailable {
			names = append(names, vm.Name)
		}
	}
	return names
}
