// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// OperatorDevice represents a device registered with the cuttlefish-operator
type OperatorDevice struct {
	DeviceId  string `json:"device_id"`
	GroupName string `json:"group_name"`
	Owner     string `json:"owner"`
	Name      string `json:"name"`
	ADBPort   int    `json:"adb_port"`
}

// CVDFleetInstance represents an instance in the CVD fleet output
type CVDFleetInstance struct {
	ADBPort        int    `json:"adb_port"`
	ADBSerial      string `json:"adb_serial"`
	InstanceName   string `json:"instance_name"`
	Status         string `json:"status"`
	WebRTCDeviceID string `json:"webrtc_device_id"`
}

// CVDFleetGroup represents a group in the CVD fleet output
type CVDFleetGroup struct {
	GroupName string             `json:"group_name"`
	Instances []CVDFleetInstance `json:"instances"`
}

// CVDFleetOutput represents the output of 'cvd fleet' command
type CVDFleetOutput struct {
	Groups []CVDFleetGroup `json:"groups"`
}

// GetCVDFleet retrieves the current CVD fleet state from the host
func (c *Client) GetCVDFleet(ctx context.Context) (*CVDFleetOutput, error) {
	cmd := fmt.Sprintf("HOME=%s %s fleet --json 2>&1 || %s fleet 2>&1",
		c.config.CVDHome, c.config.CVDPath, c.config.CVDPath)

	output, err := c.runShellScript(ctx, cmd)
	if err != nil {
		// Try to parse anyway - cvd fleet may return non-zero but still output valid JSON
		log.Debugf("cvd fleet returned error: %v, trying to parse output anyway", err)
	}

	// Extract JSON from output (cvd fleet may have log lines before JSON)
	jsonStart := -1
	for i, ch := range output {
		if ch == '{' {
			jsonStart = i
			break
		}
	}

	if jsonStart == -1 {
		return &CVDFleetOutput{Groups: []CVDFleetGroup{}}, nil
	}

	jsonOutput := output[jsonStart:]

	var fleet CVDFleetOutput
	if err := json.Unmarshal([]byte(jsonOutput), &fleet); err != nil {
		log.Warnf("Failed to parse cvd fleet output: %v (output: %s)", err, jsonOutput)
		return &CVDFleetOutput{Groups: []CVDFleetGroup{}}, nil
	}

	return &fleet, nil
}

// GetCVDInstanceNumbers returns all instance numbers currently registered in CVD
func (c *Client) GetCVDInstanceNumbers(ctx context.Context) (map[int]string, error) {
	fleet, err := c.GetCVDFleet(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[int]string)
	// Pattern to extract instance number from group name like "cvd_1", "cvd_2"
	groupPattern := regexp.MustCompile(`cvd_(\d+)`)

	for _, group := range fleet.Groups {
		matches := groupPattern.FindStringSubmatch(group.GroupName)
		if len(matches) >= 2 {
			instanceNum, err := strconv.Atoi(matches[1])
			if err == nil {
				result[instanceNum] = group.GroupName
			}
		}
	}

	return result, nil
}

// SyncCVDState synchronizes the CVD state with the runner's known sandboxes
// This removes CVD instances that are not tracked by the runner
func (c *Client) SyncCVDState(ctx context.Context) error {
	log.Info("Synchronizing CVD state with runner state...")

	// Get actual CVD instances from the host
	cvdInstances, err := c.GetCVDInstanceNumbers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get CVD fleet: %w", err)
	}

	if len(cvdInstances) == 0 {
		log.Info("No CVD instances found on host")
		return nil
	}

	log.Infof("Found %d CVD instances on host: %v", len(cvdInstances), cvdInstances)

	// Get runner's known instance numbers
	c.mutex.RLock()
	knownInstances := make(map[int]string)
	for num, sandboxId := range c.instanceNums {
		knownInstances[num] = sandboxId
	}
	c.mutex.RUnlock()

	log.Infof("Runner knows about %d instances: %v", len(knownInstances), knownInstances)

	// Find CVD instances that runner doesn't know about (orphaned)
	orphaned := []int{}
	for instanceNum, groupName := range cvdInstances {
		if _, known := knownInstances[instanceNum]; !known {
			log.Warnf("Found orphaned CVD instance %d (group: %s) - will remove", instanceNum, groupName)
			orphaned = append(orphaned, instanceNum)
		}
	}

	// Remove orphaned CVD instances
	for _, instanceNum := range orphaned {
		log.Infof("Removing orphaned CVD instance %d", instanceNum)
		if err := c.forceRemoveCVDInstance(ctx, instanceNum); err != nil {
			log.Warnf("Failed to remove orphaned instance %d: %v", instanceNum, err)
		}
	}

	if len(orphaned) > 0 {
		log.Infof("Removed %d orphaned CVD instances", len(orphaned))
	} else {
		log.Info("No orphaned CVD instances found")
	}

	return nil
}

// forceRemoveCVDInstance forcefully removes a CVD instance by its number
func (c *Client) forceRemoveCVDInstance(ctx context.Context, instanceNum int) error {
	// First try to stop the instance gracefully
	stopCmd := fmt.Sprintf("HOME=%s %s stop --instance_nums=%d 2>&1 || true",
		c.config.CVDHome, c.config.CVDPath, instanceNum)
	c.runShellScript(ctx, stopCmd)

	// Remove from CVD using group name
	groupName := fmt.Sprintf("cvd_%d", instanceNum)
	rmCmd := fmt.Sprintf("HOME=%s %s rm --group_name=%s 2>&1 || true",
		c.config.CVDHome, c.config.CVDPath, groupName)
	c.runShellScript(ctx, rmCmd)

	// Kill any processes associated with this instance
	killCmd := fmt.Sprintf(
		"pkill -9 -f 'instance_nums.*%d|CUTTLEFISH_INSTANCE=%d|cvd-%d' 2>/dev/null || true",
		instanceNum, instanceNum, instanceNum)
	c.runShellScript(ctx, killCmd)

	// Clean up temp directories
	cleanupCmd := fmt.Sprintf(
		"rm -rf /tmp/cf_avd_*/%d /tmp/cf_env_*/%d 2>/dev/null || true",
		instanceNum, instanceNum)
	c.runShellScript(ctx, cleanupCmd)

	return nil
}

// EnsureInstanceAvailable ensures an instance number is available before use
// If the instance is registered in CVD but not in the runner, it will be removed
func (c *Client) EnsureInstanceAvailable(ctx context.Context, instanceNum int) error {
	// Check if CVD has this instance registered
	cvdInstances, err := c.GetCVDInstanceNumbers(ctx)
	if err != nil {
		log.Warnf("Failed to check CVD state: %v (continuing anyway)", err)
		return nil
	}

	if groupName, exists := cvdInstances[instanceNum]; exists {
		// CVD has this instance - check if runner knows about it
		c.mutex.RLock()
		_, runnerKnows := c.instanceNums[instanceNum]
		c.mutex.RUnlock()

		if !runnerKnows {
			// Runner doesn't know about this instance - it's stale, remove it
			log.Warnf("Instance %d (group: %s) exists in CVD but not in runner - removing stale instance",
				instanceNum, groupName)
			if err := c.forceRemoveCVDInstance(ctx, instanceNum); err != nil {
				return fmt.Errorf("failed to remove stale CVD instance %d: %w", instanceNum, err)
			}
			log.Infof("Successfully removed stale CVD instance %d", instanceNum)
		}
	}

	// Also clean stale sockets/env dirs for this instance (prevents run_cvd crash)
	cleanSocketCmd := fmt.Sprintf(
		"rm -rf /tmp/cf_avd_*/cvd-%d /tmp/cf_env_*/env-%d 2>/dev/null || true",
		instanceNum, instanceNum)
	_, _ = c.runShellScript(ctx, cleanSocketCmd)

	// Also check if the device is registered in the operator (stale WebRTC registration)
	if err := c.EnsureOperatorDeviceClean(ctx, instanceNum); err != nil {
		log.Warnf("Failed to clean operator device registration: %v", err)
		// Don't fail - the create might still work
	}

	return nil
}

// GetOperatorDevices queries the cuttlefish-operator for all registered devices
func (c *Client) GetOperatorDevices(ctx context.Context) ([]OperatorDevice, error) {
	// Operator runs on port 1443 (HTTPS)
	operatorPort := 1443

	// Build target host
	var targetHost string
	if c.IsRemote() {
		targetHost = c.SSHHost
		if idx := strings.Index(targetHost, "@"); idx != -1 {
			targetHost = targetHost[idx+1:]
		}
	} else {
		targetHost = "localhost"
	}

	devicesURL := fmt.Sprintf("https://%s:%d/devices", targetHost, operatorPort)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Get(devicesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query operator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("operator returned status %d", resp.StatusCode)
	}

	var devices []OperatorDevice
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("failed to decode devices: %w", err)
	}

	return devices, nil
}

// EnsureOperatorDeviceClean checks if a device with the given instance number
// is registered with the cuttlefish-operator and removes stale registrations
func (c *Client) EnsureOperatorDeviceClean(ctx context.Context, instanceNum int) error {
	devices, err := c.GetOperatorDevices(ctx)
	if err != nil {
		log.Debugf("Could not query operator devices: %v (continuing anyway)", err)
		return nil
	}

	// Device IDs follow the pattern: cvd_{group}-{instance}-{instance}
	// We need to check if any device with our instance number exists
	instanceStr := fmt.Sprintf("-%d-", instanceNum)

	for _, dev := range devices {
		// Check if this device ID contains our instance number
		if strings.Contains(dev.DeviceId, instanceStr) || dev.Name == fmt.Sprintf("%d", instanceNum) {
			// Check if runner knows about this instance
			c.mutex.RLock()
			_, runnerKnows := c.instanceNums[instanceNum]
			c.mutex.RUnlock()

			if !runnerKnows {
				log.Warnf("Device %s (instance %d) is registered in operator but runner doesn't know about it - cleaning up",
					dev.DeviceId, instanceNum)

				// The operator doesn't have a direct unregister API
				// We need to either restart the operator or let cvd rm handle it
				// For now, try cvd rm for the group
				if dev.GroupName != "" {
					rmCmd := fmt.Sprintf(
						"HOME=%s %s rm --group_name=%s 2>/dev/null || true",
						c.config.CVDHome, c.config.CVDPath, dev.GroupName,
					)
					_, _ = c.runShellScript(ctx, rmCmd)
					log.Infof("Attempted to remove stale CVD group %s", dev.GroupName)
				}
			}
		}
	}

	return nil
}
