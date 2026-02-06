// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"fmt"
	"sync"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	runnerapiclient "github.com/daytonaio/runner-android/pkg/apiclient"
	log "github.com/sirupsen/logrus"
)

// HealthMonitor monitors CVD instance health and reports crashes to the main API
type HealthMonitor struct {
	client      *Client
	apiClient   *apiclient.APIClient
	interval    time.Duration
	mutex       sync.Mutex
	lastStates  map[string]InstanceState // Track last known states to avoid duplicate reports
	crashCounts map[string]int           // Track crash counts for each sandbox
	maxRetries  int                      // Max retry attempts before reporting error
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// HealthMonitorConfig configures the health monitor
type HealthMonitorConfig struct {
	Interval   time.Duration // How often to check (default: 30s)
	MaxRetries int           // Max retries before reporting crash (default: 2)
}

// NewHealthMonitor creates a new CVD health monitor
func NewHealthMonitor(client *Client, cfg *HealthMonitorConfig) (*HealthMonitor, error) {
	apiClient, err := runnerapiclient.GetApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	interval := 30 * time.Second
	if cfg != nil && cfg.Interval > 0 {
		interval = cfg.Interval
	}

	maxRetries := 2
	if cfg != nil && cfg.MaxRetries > 0 {
		maxRetries = cfg.MaxRetries
	}

	return &HealthMonitor{
		client:      client,
		apiClient:   apiClient,
		interval:    interval,
		lastStates:  make(map[string]InstanceState),
		crashCounts: make(map[string]int),
		maxRetries:  maxRetries,
		stopCh:      make(chan struct{}),
	}, nil
}

// Start begins the health monitoring loop
func (m *HealthMonitor) Start(ctx context.Context) {
	m.wg.Add(1)
	go m.monitorLoop(ctx)
	log.Info("CVD health monitor started")
}

// Stop stops the health monitor
func (m *HealthMonitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
	log.Info("CVD health monitor stopped")
}

// monitorLoop runs the periodic health check
func (m *HealthMonitor) monitorLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Run initial check
	m.checkHealth(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkHealth(ctx)
		}
	}
}

// checkHealth checks all tracked sandboxes and reports crashes
func (m *HealthMonitor) checkHealth(ctx context.Context) {
	// Get CVD fleet status
	fleet, err := m.client.GetCVDFleet(ctx)
	if err != nil {
		log.Warnf("Health monitor: failed to get CVD fleet: %v", err)
		return
	}

	// Build map of CVD instance statuses
	cvdStatuses := make(map[int]string) // instanceNum -> status
	for _, group := range fleet.Groups {
		for _, instance := range group.Instances {
			// Extract instance number from group name (e.g., "cvd_1" -> 1)
			var instanceNum int
			if _, err := fmt.Sscanf(group.GroupName, "cvd_%d", &instanceNum); err == nil {
				cvdStatuses[instanceNum] = instance.Status
			}
		}
	}

	// Check each tracked sandbox
	m.client.mutex.RLock()
	instances := make(map[string]*InstanceInfo)
	for sandboxId, info := range m.client.instances {
		instances[sandboxId] = info
	}
	m.client.mutex.RUnlock()

	for sandboxId, info := range instances {
		m.checkSandboxHealth(ctx, sandboxId, info, cvdStatuses)
	}
}

// checkSandboxHealth checks a single sandbox's health
func (m *HealthMonitor) checkSandboxHealth(ctx context.Context, sandboxId string, info *InstanceInfo, cvdStatuses map[int]string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get CVD status for this instance
	cvdStatus, exists := cvdStatuses[info.InstanceNum]

	// Determine current state
	var currentState InstanceState
	if !exists {
		// Instance not in CVD fleet at all - it's crashed/removed
		currentState = InstanceStateStopped
	} else if cvdStatus == "Running" {
		currentState = InstanceStateRunning
		// Reset crash count on successful running state
		delete(m.crashCounts, sandboxId)
	} else {
		// CVD reports stopped/other status
		currentState = InstanceStateStopped
	}

	// Get last known state
	lastState, hasLastState := m.lastStates[sandboxId]

	// Check for state transitions that indicate a crash
	if hasLastState && lastState == InstanceStateRunning && currentState == InstanceStateStopped {
		// Instance was running but now stopped - potential crash
		m.crashCounts[sandboxId]++
		crashCount := m.crashCounts[sandboxId]

		log.Warnf("Health monitor: sandbox %s (instance %d) may have crashed (count: %d/%d)",
			sandboxId, info.InstanceNum, crashCount, m.maxRetries)

		if crashCount >= m.maxRetries {
			// Confirmed crash - report to API
			log.Errorf("Health monitor: sandbox %s confirmed crashed, reporting to API", sandboxId)
			m.reportCrash(ctx, sandboxId, "CVD instance stopped unexpectedly")
			delete(m.crashCounts, sandboxId)
		}
	} else if !hasLastState && info.State == InstanceStateRunning && currentState == InstanceStateStopped {
		// First check and instance should be running but isn't
		m.crashCounts[sandboxId]++
		crashCount := m.crashCounts[sandboxId]

		log.Warnf("Health monitor: sandbox %s (instance %d) not running as expected (count: %d/%d)",
			sandboxId, info.InstanceNum, crashCount, m.maxRetries)

		if crashCount >= m.maxRetries {
			log.Errorf("Health monitor: sandbox %s not running, reporting to API", sandboxId)
			m.reportCrash(ctx, sandboxId, "CVD instance not running")
			delete(m.crashCounts, sandboxId)
		}
	}

	// Update last state
	m.lastStates[sandboxId] = currentState
}

// reportCrash reports a crash to the main API
func (m *HealthMonitor) reportCrash(ctx context.Context, sandboxId string, reason string) {
	// Create update request
	updateDto := apiclient.NewUpdateSandboxStateDto("error")
	updateDto.SetErrorReason(reason)

	// Send to API
	req := m.apiClient.SandboxAPI.UpdateSandboxState(ctx, sandboxId).
		UpdateSandboxStateDto(*updateDto)

	_, err := req.Execute()
	if err != nil {
		log.Errorf("Health monitor: failed to report crash for sandbox %s: %v", sandboxId, err)
		return
	}

	log.Infof("Health monitor: reported crash for sandbox %s to API", sandboxId)

	// Update local state
	m.client.mutex.Lock()
	if info, exists := m.client.instances[sandboxId]; exists {
		info.State = InstanceStateStopped
	}
	m.client.mutex.Unlock()
}

// ClearSandbox removes a sandbox from monitoring (called when sandbox is destroyed)
func (m *HealthMonitor) ClearSandbox(sandboxId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.lastStates, sandboxId)
	delete(m.crashCounts, sandboxId)
}

// ResetSandboxState resets the state tracking for a sandbox (called when starting)
func (m *HealthMonitor) ResetSandboxState(sandboxId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.lastStates, sandboxId)
	delete(m.crashCounts, sandboxId)
}
