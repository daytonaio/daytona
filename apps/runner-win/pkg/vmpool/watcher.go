// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package vmpool

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/libvirt"
	log "github.com/sirupsen/logrus"
)

// Watcher monitors the pool and creates new VMs to maintain the target size
type Watcher struct {
	pool    *Pool
	libvirt *libvirt.LibVirt
	config  PoolConfig

	// createMu ensures only one VM is created at a time
	createMu sync.Mutex
	// stopCh signals the watcher to stop
	stopCh chan struct{}
}

// NewWatcher creates a new pool watcher
func NewWatcher(pool *Pool, libvirtClient *libvirt.LibVirt, config PoolConfig) *Watcher {
	return &Watcher{
		pool:    pool,
		libvirt: libvirtClient,
		config:  config,
		stopCh:  make(chan struct{}),
	}
}

// Start begins watching the pool and replenishing VMs as needed
func (w *Watcher) Start(ctx context.Context) {
	if !w.config.IsEnabled() {
		log.Info("VM pool is disabled (size=0), watcher not starting")
		return
	}

	log.Infof("Starting VM pool watcher (target size: %d, interval: %v)", w.config.Size, w.config.WatchInterval)

	// First, try to recover existing pool VMs from a previous run
	if err := w.pool.RecoverExistingPoolVMs(ctx); err != nil {
		log.Warnf("Failed to recover existing pool VMs: %v", err)
	}

	// Update the next index counter to avoid name collisions
	w.pool.UpdateNextIndex()

	// Log current pool status after recovery
	stats := w.pool.Stats()
	log.Infof("Pool status after recovery: available=%d, claimed=%d, creating=%d, target=%d",
		stats.Available, stats.Claimed, stats.Creating, stats.TargetSize)

	// Initial population of the pool (only if needed after recovery)
	go w.populatePool(ctx)

	ticker := time.NewTicker(w.config.WatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("VM pool watcher stopping (context cancelled)")
			return
		case <-w.stopCh:
			log.Info("VM pool watcher stopping (stop signal received)")
			return
		case <-ticker.C:
			w.checkAndReplenish(ctx)
		}
	}
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	close(w.stopCh)
}

// populatePool creates VMs until the pool reaches target size
func (w *Watcher) populatePool(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !w.pool.NeedsReplenishment() {
			stats := w.pool.Stats()
			log.Infof("Pool is at target size (available: %d, creating: %d, target: %d)",
				stats.Available, stats.Creating, stats.TargetSize)
			return
		}

		if err := w.createPoolVM(ctx); err != nil {
			log.Errorf("Failed to create pool VM: %v", err)
			// Wait a bit before retrying
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}
}

// checkAndReplenish checks the pool size and creates VMs if needed
func (w *Watcher) checkAndReplenish(ctx context.Context) {
	stats := w.pool.Stats()
	log.Debugf("Pool status: available=%d, claimed=%d, creating=%d, target=%d",
		stats.Available, stats.Claimed, stats.Creating, stats.TargetSize)

	if w.pool.NeedsReplenishment() {
		log.Infof("Pool needs replenishment (available: %d, creating: %d, target: %d)",
			stats.Available, stats.Creating, stats.TargetSize)
		go w.createPoolVM(ctx)
	}
}

// createPoolVM creates a new VM for the pool
func (w *Watcher) createPoolVM(ctx context.Context) error {
	// Ensure only one VM is created at a time
	w.createMu.Lock()
	defer w.createMu.Unlock()

	// Double-check we still need a VM after acquiring the lock
	if !w.pool.NeedsReplenishment() {
		return nil
	}

	// Generate a unique name for the pool VM
	vmName := w.pool.GenerateVMName()
	log.Infof("Creating pool VM: %s", vmName)

	// Mark as creating to prevent duplicate creation
	w.pool.MarkCreating(vmName)

	// Create the VM using memory snapshot path
	// We create a minimal sandbox DTO just for the pool VM
	sandboxDTO := dto.CreateSandboxDTO{
		Id:           vmName,
		UserId:       "pool",
		Snapshot:     "pool",
		OsUser:       "pool",
		CpuQuota:     2,    // Minimum for Windows
		MemoryQuota:  4096, // 4GB minimum for Windows
		StorageQuota: 50,
	}

	uuid, domainName, err := w.libvirt.Create(ctx, sandboxDTO)
	if err != nil {
		w.pool.CancelCreating(vmName)
		// Check if this is a collision error - if so, the next attempt will use a new index
		if strings.Contains(err.Error(), "domain name collision") {
			log.Warnf("Pool VM %s name collision, will retry with next index: %v", vmName, err)
			return nil // Return nil so checkAndReplenish will try again immediately
		}
		return fmt.Errorf("failed to create pool VM %s: %w", vmName, err)
	}

	// IMPORTANT: Use the actual domain name returned from Create(), not vmName
	// With memory snapshot, the domain name is generated (e.g., sndbx-xxxxxxxxx)
	// but we track it in the pool using vmName (e.g., pool-vm-0000001)
	log.Infof("Pool VM %s created as domain %s (UUID: %s), waiting for daemon...", vmName, domainName, uuid)

	// CRITICAL: Wait for daemon to be ready before pausing
	// Use the sandbox ID (vmName) for lookup since LookupDomainBySandboxId handles both naming schemes
	if err := w.libvirt.WaitForDaemonReady(ctx, vmName); err != nil {
		log.Errorf("Daemon not ready on pool VM %s (domain %s), destroying: %v", vmName, domainName, err)
		// Destroy the failed VM using sandbox ID
		if destroyErr := w.libvirt.Destroy(ctx, vmName); destroyErr != nil {
			log.Errorf("Failed to destroy failed pool VM %s: %v", vmName, destroyErr)
		}
		w.pool.CancelCreating(vmName)
		return fmt.Errorf("daemon not ready on pool VM %s: %w", vmName, err)
	}

	log.Infof("Daemon ready on pool VM %s (domain %s), pausing...", vmName, domainName)

	// Pause the VM using sandbox ID
	if err := w.libvirt.Pause(ctx, vmName); err != nil {
		log.Errorf("Failed to pause pool VM %s: %v", vmName, err)
		// Destroy the VM since we can't use it
		if destroyErr := w.libvirt.Destroy(ctx, vmName); destroyErr != nil {
			log.Errorf("Failed to destroy pool VM %s after pause failure: %v", vmName, destroyErr)
		}
		w.pool.CancelCreating(vmName)
		return fmt.Errorf("failed to pause pool VM %s: %w", vmName, err)
	}

	// Get the IP address and MAC of the VM
	ip := w.libvirt.GetActualDomainIP(vmName)
	mac := libvirt.GenerateMACFromSandboxID(vmName)

	// Update pool with the finished VM
	// IMPORTANT: Store both the pool name (vmName) and the actual domain name
	w.pool.FinishCreating(vmName, domainName, uuid, ip, mac)
	log.Infof("Pool VM %s (domain: %s) is ready and paused", vmName, domainName)

	return nil
}

// GetPool returns the underlying pool
func (w *Watcher) GetPool() *Pool {
	return w.pool
}
