// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// VMMemoryStats holds memory statistics for a VM collected from the daemon
type VMMemoryStats struct {
	SandboxId       string // Sandbox ID
	MaxMemoryKiB    uint64 // Maximum memory from VM config
	BalloonSizeKiB  uint64 // Current balloon inflation (memory reclaimed)
	MemTotalKiB     uint64 // Total memory visible to guest
	MemFreeKiB      uint64 // Free memory in guest
	MemAvailableKiB uint64 // Memory available without swapping
	BuffersKiB      uint64 // Buffer memory
	CachedKiB       uint64 // Cached memory
	LastUpdate      int64  // Unix timestamp of when stats were collected
}

// UsedMemoryKiB returns the memory actually being used by the guest
func (s *VMMemoryStats) UsedMemoryKiB() uint64 {
	if s.MemAvailableKiB > s.MemTotalKiB {
		return 0
	}
	return s.MemTotalKiB - s.MemAvailableKiB
}

// IsBalloonDriverActive returns true if the daemon is responding with valid stats
// For Cloud Hypervisor, we detect this by checking if we got valid memory info
func (s *VMMemoryStats) IsBalloonDriverActive() bool {
	// If we have valid memory stats from the daemon, the balloon driver is working
	return s.MemTotalKiB > 0 && s.LastUpdate > 0
}

// daemonMemoryStatsResponse matches the daemon's /memory-stats response
type daemonMemoryStatsResponse struct {
	MemTotalKiB     uint64 `json:"memTotalKiB"`
	MemFreeKiB      uint64 `json:"memFreeKiB"`
	MemAvailableKiB uint64 `json:"memAvailableKiB"`
	BuffersKiB      uint64 `json:"buffersKiB"`
	CachedKiB       uint64 `json:"cachedKiB"`
}

// GetVMMemoryStats retrieves memory statistics for a single VM by querying the daemon
func (c *Client) GetVMMemoryStats(ctx context.Context, sandboxId string) (*VMMemoryStats, error) {
	// Get sandbox info to find IP and memory config
	sandboxInfo, err := c.GetSandboxInfo(ctx, sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox info: %w", err)
	}

	if sandboxInfo.State != VmStateRunning {
		return nil, fmt.Errorf("VM is not running (state: %s)", sandboxInfo.State)
	}

	if sandboxInfo.IpAddress == "" {
		return nil, fmt.Errorf("VM has no IP address")
	}

	// Query daemon's /memory-stats endpoint (must run inside network namespace)
	daemonStats, err := c.queryDaemonMemoryStats(ctx, sandboxId, sandboxInfo.IpAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to query daemon memory stats: %w", err)
	}

	// Get VM info for max memory and current balloon size
	vmInfo, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info: %w", err)
	}

	stats := &VMMemoryStats{
		SandboxId:       sandboxId,
		MemTotalKiB:     daemonStats.MemTotalKiB,
		MemFreeKiB:      daemonStats.MemFreeKiB,
		MemAvailableKiB: daemonStats.MemAvailableKiB,
		BuffersKiB:      daemonStats.BuffersKiB,
		CachedKiB:       daemonStats.CachedKiB,
		LastUpdate:      time.Now().Unix(),
	}

	// Get max memory from VM config
	if vmInfo.Config != nil && vmInfo.Config.Memory != nil {
		stats.MaxMemoryKiB = vmInfo.Config.Memory.Size / 1024 // bytes to KiB
	}

	// Get current balloon size from VM config
	if vmInfo.Config != nil && vmInfo.Config.Balloon != nil {
		stats.BalloonSizeKiB = vmInfo.Config.Balloon.Size / 1024 // bytes to KiB
	}

	return stats, nil
}

// GetAllVMMemoryStats retrieves memory statistics for all running VMs
func (c *Client) GetAllVMMemoryStats(ctx context.Context) (map[string]*VMMemoryStats, error) {
	// Get list of all sandboxes
	sandboxIds, err := c.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sandboxes: %w", err)
	}

	stats := make(map[string]*VMMemoryStats)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Query stats in parallel for better performance
	for _, sandboxId := range sandboxIds {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			vmStats, err := c.GetVMMemoryStats(ctx, id)
			if err != nil {
				log.Debugf("Failed to get memory stats for %s: %v", id, err)
				return
			}

			mu.Lock()
			stats[id] = vmStats
			mu.Unlock()
		}(sandboxId)
	}

	wg.Wait()
	return stats, nil
}

// queryDaemonMemoryStats queries the daemon's /memory-stats endpoint
// The command must run inside the sandbox's network namespace to reach the VM's IP
func (c *Client) queryDaemonMemoryStats(ctx context.Context, sandboxId, vmIP string) (*daemonMemoryStatsResponse, error) {
	// Build curl command
	curlCmd := fmt.Sprintf("curl -s --connect-timeout 5 --max-time 10 http://%s:2280/memory-stats", vmIP)

	// Run curl inside the sandbox's network namespace
	// The VM IP (192.168.0.2) is only reachable from inside the namespace
	output, err := c.netnsPool.ExecInNamespace(ctx, sandboxId, curlCmd)
	if err != nil {
		return nil, fmt.Errorf("curl failed: %w", err)
	}

	var stats daemonMemoryStatsResponse
	if err := json.Unmarshal([]byte(output), &stats); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (output: %s)", err, output)
	}

	return &stats, nil
}
