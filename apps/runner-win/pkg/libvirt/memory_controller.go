// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// MemoryControllerConfig holds configuration for the memory controller
type MemoryControllerConfig struct {
	LibVirt           *LibVirt
	StatsStore        *StatsStore   // Optional: for recording memory stats history
	CheckInterval     time.Duration // How often to check and rebalance (default: 30s)
	MinVMMemoryGB     uint64        // Minimum memory per VM in GB (default: 4)
	SafetyBufferGB    uint64        // Minimum headroom per VM in GB (default: 2)
	SafetyBufferRatio float64       // Headroom as ratio of used memory (default: 0.25)
	Enabled           bool          // Whether ballooning is enabled
}

// MemoryController manages aggressive memory ballooning for all VMs
// It periodically reclaims unused memory from VMs to maximize host free memory,
// which improves availability scoring for VM scheduling.
type MemoryController struct {
	config    MemoryControllerConfig
	logger    *log.Entry
	lastStats map[string]*VMMemoryStats
	statsMu   sync.RWMutex
	running   bool
	runMu     sync.Mutex
}

// NewMemoryController creates a new memory controller with the given configuration
func NewMemoryController(config MemoryControllerConfig) *MemoryController {
	// Set defaults
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.MinVMMemoryGB == 0 {
		config.MinVMMemoryGB = 4
	}
	if config.SafetyBufferGB == 0 {
		config.SafetyBufferGB = 2
	}
	if config.SafetyBufferRatio == 0 {
		config.SafetyBufferRatio = 0.25
	}

	return &MemoryController{
		config:    config,
		logger:    log.WithField("component", "memory_controller"),
		lastStats: make(map[string]*VMMemoryStats),
	}
}

// Start begins the memory controller background loop
func (mc *MemoryController) Start(ctx context.Context) {
	mc.runMu.Lock()
	if mc.running {
		mc.runMu.Unlock()
		return
	}
	mc.running = true
	mc.runMu.Unlock()

	mc.logger.Infof("Starting memory controller (interval: %v, min VM: %dGB, buffer: %dGB or %.0f%%)",
		mc.config.CheckInterval, mc.config.MinVMMemoryGB, mc.config.SafetyBufferGB, mc.config.SafetyBufferRatio*100)

	ticker := time.NewTicker(mc.config.CheckInterval)
	defer ticker.Stop()

	// Run immediately on start
	mc.balanceAllVMs(ctx)

	for {
		select {
		case <-ctx.Done():
			mc.logger.Info("Memory controller stopped")
			mc.runMu.Lock()
			mc.running = false
			mc.runMu.Unlock()
			return
		case <-ticker.C:
			mc.balanceAllVMs(ctx)
		}
	}
}

// balanceAllVMs collects stats for all VMs and adjusts their memory
func (mc *MemoryController) balanceAllVMs(ctx context.Context) {
	startTime := time.Now()

	// Get memory stats for all running VMs
	stats, err := mc.config.LibVirt.GetAllVMMemoryStats(ctx)
	if err != nil {
		mc.logger.Warnf("Failed to get VM memory stats: %v", err)
		return
	}

	if len(stats) == 0 {
		mc.logger.Debug("No running VMs to balance")
		return
	}

	// Store stats for later reference
	mc.statsMu.Lock()
	mc.lastStats = stats
	mc.statsMu.Unlock()

	// Record stats to persistent storage (async, non-blocking)
	if mc.config.StatsStore != nil {
		mc.config.StatsStore.RecordBatch(stats)
	}

	var (
		totalReclaimed uint64
		totalReturned  uint64
		vmsBallooned   int
		vmsExpanded    int
		vmsSkipped     int
	)

	for vmName, vmStats := range stats {
		// Skip VMs without active balloon driver
		if !vmStats.IsBalloonDriverActive() {
			mc.logger.Debugf("Skipping %s: balloon driver not active (last_update=0)", vmName)
			vmsSkipped++
			continue
		}

		// Calculate target memory
		targetKiB := mc.calculateTargetMemory(vmStats)

		// Determine action
		currentKiB := vmStats.ActualKiB
		if targetKiB == currentKiB {
			// No change needed
			continue
		}

		// Apply the change
		if err := mc.config.LibVirt.SetVMMemoryByName(ctx, vmName, targetKiB); err != nil {
			mc.logger.Warnf("Failed to set memory for %s: %v", vmName, err)
			continue
		}

		if targetKiB < currentKiB {
			// Reclaimed memory (balloon inflated)
			reclaimedKiB := currentKiB - targetKiB
			totalReclaimed += reclaimedKiB
			vmsBallooned++
			mc.logger.Infof("Ballooned %s: %s -> %s (reclaimed %s, unused was %s)",
				vmName,
				formatMemory(currentKiB),
				formatMemory(targetKiB),
				formatMemory(reclaimedKiB),
				formatMemory(vmStats.UnusedKiB))
		} else {
			// Returned memory (balloon deflated)
			returnedKiB := targetKiB - currentKiB
			totalReturned += returnedKiB
			vmsExpanded++
			mc.logger.Infof("Expanded %s: %s -> %s (returned %s)",
				vmName,
				formatMemory(currentKiB),
				formatMemory(targetKiB),
				formatMemory(returnedKiB))
		}
	}

	elapsed := time.Since(startTime)

	// Log summary
	if vmsBallooned > 0 || vmsExpanded > 0 {
		mc.logger.Infof("Cycle complete in %v: %d VMs processed, %d ballooned (-%s), %d expanded (+%s), %d skipped",
			elapsed, len(stats), vmsBallooned, formatMemory(totalReclaimed),
			vmsExpanded, formatMemory(totalReturned), vmsSkipped)
	} else {
		mc.logger.Debugf("Cycle complete in %v: %d VMs checked, no changes needed, %d skipped",
			elapsed, len(stats), vmsSkipped)
	}
}

// calculateTargetMemory determines the optimal memory allocation for a VM
// Formula: target = used + buffer, where buffer = max(SafetyBufferGB, used * SafetyBufferRatio)
func (mc *MemoryController) calculateTargetMemory(stats *VMMemoryStats) uint64 {
	// Calculate used memory (actual - unused)
	usedKiB := stats.UsedMemoryKiB()

	// Calculate safety buffer
	// Use the larger of: fixed buffer OR percentage of used
	fixedBufferKiB := mc.config.SafetyBufferGB * 1024 * 1024 // GB to KiB
	ratioBufferKiB := uint64(float64(usedKiB) * mc.config.SafetyBufferRatio)

	bufferKiB := fixedBufferKiB
	if ratioBufferKiB > bufferKiB {
		bufferKiB = ratioBufferKiB
	}

	// Target = used + buffer
	targetKiB := usedKiB + bufferKiB

	// Clamp to minimum VM memory
	minKiB := mc.config.MinVMMemoryGB * 1024 * 1024 // GB to KiB
	if targetKiB < minKiB {
		targetKiB = minKiB
	}

	// Clamp to maximum (can't exceed VM's max memory)
	if targetKiB > stats.MaxMemoryKiB {
		targetKiB = stats.MaxMemoryKiB
	}

	return targetKiB
}

// GetLastStats returns the most recent memory stats for all VMs
func (mc *MemoryController) GetLastStats() map[string]*VMMemoryStats {
	mc.statsMu.RLock()
	defer mc.statsMu.RUnlock()

	// Return a copy to avoid race conditions
	copy := make(map[string]*VMMemoryStats, len(mc.lastStats))
	for k, v := range mc.lastStats {
		statsCopy := *v
		copy[k] = &statsCopy
	}
	return copy
}

// formatMemory formats memory in KiB to a human-readable string
func formatMemory(kib uint64) string {
	if kib >= 1024*1024 {
		return fmt.Sprintf("%.1fGB", float64(kib)/(1024*1024))
	}
	if kib >= 1024 {
		return fmt.Sprintf("%.1fMB", float64(kib)/1024)
	}
	return fmt.Sprintf("%dKiB", kib)
}
