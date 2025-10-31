// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type DockerMonitor struct {
	apiClient       client.APIClient
	ctx             context.Context
	cancel          context.CancelFunc
	netRulesManager *netrules.NetRulesManager
}

func NewDockerMonitor(apiClient client.APIClient, netRulesManager *netrules.NetRulesManager) *DockerMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &DockerMonitor{
		apiClient:       apiClient,
		ctx:             ctx,
		cancel:          cancel,
		netRulesManager: netRulesManager,
	}
}

func (dm *DockerMonitor) Stop() {
	dm.cancel()
}

func (dm *DockerMonitor) Start() error {

	slog.Info("Starting Docker monitor")

	// Start periodic reconciliation
	go dm.reconcilerLoop()

	// Main monitoring loop
	for {
		select {
		case <-dm.ctx.Done():
			slog.Info("Context cancelled, stopping monitor...")
			return dm.ctx.Err()

		default:
			if err := dm.monitorEvents(); err != nil {
				if isConnectionError(err) {
					slog.Warn("Events stream ended", "error", err)
					slog.Info("Reopening events stream in 2 seconds...")
					time.Sleep(2 * time.Second)
					continue
				} else {
					slog.Error("Fatal error in monitoring", "error", err)
					return err
				}
			}
		}
	}
}

// isConnectionError checks if the error is related to connection loss
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// io.EOF is the normal way the Docker Events stream ends
	if err == io.EOF {
		return true
	}

	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "unexpected EOF") ||
		strings.Contains(errStr, "Cannot connect to the Docker daemon")
}

// monitorEvents handles the actual event monitoring with proper error handling
func (dm *DockerMonitor) monitorEvents() error {
	// Create event filters to monitor only container create and stop events
	eventFilters := events.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
			filters.Arg("event", "start"),
			filters.Arg("event", "stop"),
			filters.Arg("event", "kill"),
			filters.Arg("event", "destroy"),
		),
	}

	// Start listening for events
	eventsChan, errsChan := dm.apiClient.Events(dm.ctx, eventFilters)

	// Reconnection established successfully
	dm.reconcileNetworkRules("filter", "DOCKER-USER")
	dm.reconcileNetworkRules("mangle", "PREROUTING")

	for {
		select {
		case event := <-eventsChan:
			slog.Debug("Received event", "event", event)
			dm.handleContainerEvent(event)

		case err := <-errsChan:
			if err != nil {
				slog.Warn("Events stream ended", "error", err)
				return err
			}

		case <-dm.ctx.Done():
			return dm.ctx.Err()
		}
	}
}

func (dm *DockerMonitor) handleContainerEvent(event events.Message) {
	containerID := event.Actor.ID
	action := event.Action

	switch action {
	case "start":
		ct, err := dm.apiClient.ContainerInspect(dm.ctx, containerID)
		if err != nil {
			slog.Error("Error inspecting container", "error", err)
			return
		}
		shortContainerID := containerID[:12]
		err = dm.netRulesManager.AssignNetworkRules(shortContainerID, ct.NetworkSettings.IPAddress)
		if err != nil {
			slog.Error("Error assigning network rules", "error", err)
		}
	case "stop":
	case "kill":
		shortContainerID := containerID[:12]
		err := dm.netRulesManager.UnassignNetworkRules(shortContainerID)
		if err != nil {
			slog.Error("Error unassigning network rules", "error", err)
		}
		err = dm.netRulesManager.RemoveNetworkLimiter(shortContainerID)
		if err != nil {
			slog.Error("Error removing network limiter", "error", err)
		}
	case "destroy":
		shortContainerID := containerID[:12]
		err := dm.netRulesManager.DeleteNetworkRules(shortContainerID)
		if err != nil {
			slog.Error("Error deleting network rules", "error", err)
		}
	}
}

// reconcileNetworkRules is called when reconnection is established
func (dm *DockerMonitor) reconcileNetworkRules(table string, chain string) {
	// List all DOCKER-USER rules that jump to Daytona chains
	rules, err := dm.netRulesManager.ListDaytonaRules(table, chain)
	if err != nil {
		slog.Error("Error listing Daytona rules", "error", err)
		return
	}

	for _, rule := range rules {
		// Parse the rule to extract chain name and source IP
		args, err := netrules.ParseRuleArguments(rule)
		if err != nil {
			slog.Error("Error parsing rule", "rule", rule, "error", err)
			continue
		}

		// Find the chain name and source IP from the rule arguments
		var chainName, sourceIP string
		for i, arg := range args {
			if arg == "-j" && i+1 < len(args) {
				chainName = args[i+1]
			}
			if arg == "-s" && i+1 < len(args) {
				sourceIP = args[i+1]
			}
		}

		if chainName == "" || sourceIP == "" {
			slog.Warn("Could not extract chain name or source IP from rule", "rule", rule)
			continue
		}

		// Extract container ID from chain name (remove DAYTONA-SB- prefix)
		containerID := strings.TrimPrefix(chainName, "DAYTONA-SB-")
		if containerID == chainName {
			slog.Warn("Invalid chain name format", "chainName", chainName)
			continue
		}

		// Inspect the container to get its current IP
		container, err := dm.apiClient.ContainerInspect(dm.ctx, containerID)
		if err != nil {
			slog.Error("Error inspecting container", "containerID", containerID, "error", err)
			// Container doesn't exist, unassign the rules
			if err := dm.netRulesManager.UnassignNetworkRules(containerID); err != nil {
				slog.Error("Error unassigning rules for non-existent container", "containerID", containerID, "error", err)
			} else {
				slog.Info("Unassigned rules for non-existent container", "containerID", containerID)
			}
			continue
		}

		// Check if the container IP matches the rule's source IP
		// Handle CIDR notation by extracting just the IP part
		ruleIP := sourceIP
		if strings.Contains(sourceIP, "/") {
			ruleIP = strings.Split(sourceIP, "/")[0]
		}

		if container.NetworkSettings.IPAddress != ruleIP {
			slog.Warn("IP mismatch for container", "containerID", containerID, "ruleIP", ruleIP, "containerIP", container.NetworkSettings.IPAddress)

			// Delete only this specific mismatched rule
			if err := dm.netRulesManager.DeleteChainRule(table, chain, rule); err != nil {
				slog.Error("Error deleting mismatched rule for container", "containerID", containerID, "error", err)
			} else {
				slog.Info("Deleted mismatched rule for container", "containerID", containerID)
			}
		}
	}
}

// reconcileChains removes orphaned chains for non-existent containers
func (dm *DockerMonitor) reconcileChains(table string) {
	// List all chains that start with DAYTONA-SB-
	chains, err := dm.netRulesManager.ListDaytonaChains(table)
	if err != nil {
		slog.Error("Error listing Daytona chains", "error", err)
		return
	}

	for _, chain := range chains {
		// Extract container ID from chain name (remove DAYTONA-SB- prefix)
		containerID := strings.TrimPrefix(chain, "DAYTONA-SB-")
		if containerID == chain {
			slog.Warn("Invalid chain name format", "chain", chain)
			continue
		}

		// Check if the container exists
		_, err := dm.apiClient.ContainerInspect(dm.ctx, containerID)
		if err != nil {
			slog.Info("Container does not exist, deleting chain", "containerID", containerID, "chain", chain)

			// Delete the orphaned chain
			if err := dm.netRulesManager.ClearAndDeleteChain(table, chain); err != nil {
				slog.Error("Error deleting orphaned chain", "chain", chain, "error", err)
			} else {
				slog.Info("Deleted orphaned chain", "chain", chain)
			}
		}
	}
}

// reconcilerLoop runs reconciliation every minute
func (dm *DockerMonitor) reconcilerLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			slog.Debug("Reconciling network rules")
			dm.reconcileNetworkRules("filter", "DOCKER-USER")
			dm.reconcileNetworkRules("mangle", "PREROUTING")
			dm.reconcileChains("filter")
			dm.reconcileChains("mangle")
		}
	}
}
