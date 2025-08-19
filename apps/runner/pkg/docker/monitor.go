// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
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

	log.Info("Starting Docker monitor")

	// Start periodic reconciliation
	go dm.reconcilerLoop()

	// Main monitoring loop
	for {
		select {
		case <-dm.ctx.Done():
			log.Info("Context cancelled, stopping monitor...")
			return dm.ctx.Err()

		default:
			if err := dm.monitorEvents(); err != nil {
				if isConnectionError(err) {
					log.Warnf("Events stream ended: %v", err)
					log.Info("Reopening events stream in 2 seconds...")
					time.Sleep(2 * time.Second)
					continue
				} else {
					log.Errorf("Fatal error in monitoring: %v", err)
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
	dm.reconcileNetworkRules()

	for {
		select {
		case event := <-eventsChan:
			log.Debug("Received event", event)
			dm.handleContainerEvent(event)

		case err := <-errsChan:
			if err != nil {
				log.Warnf("Events stream ended: %v", err)
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
			log.Errorf("Error inspecting container: %v", err)
			return
		}
		shortContainerID := containerID[:12]
		err = dm.netRulesManager.AssignNetworkRules(shortContainerID, ct.NetworkSettings.IPAddress)
		if err != nil {
			log.Errorf("Error assigning network rules: %v", err)
		}
	case "stop":
	case "kill":
		shortContainerID := containerID[:12]
		err := dm.netRulesManager.UnassignNetworkRules(shortContainerID)
		if err != nil {
			log.Errorf("Error unassigning network rules: %v", err)
		}
	case "destroy":
		shortContainerID := containerID[:12]
		err := dm.netRulesManager.DeleteNetworkRules(shortContainerID)
		if err != nil {
			log.Errorf("Error deleting network rules: %v", err)
		}
	}
}

// reconcileNetworkRules is called when reconnection is established
func (dm *DockerMonitor) reconcileNetworkRules() {
	// List all DOCKER-USER rules that jump to Daytona chains
	rules, err := dm.netRulesManager.ListDaytonaRules()
	if err != nil {
		log.Errorf("Error listing Daytona rules: %v", err)
		return
	}

	for _, rule := range rules {
		// Parse the rule to extract chain name and source IP
		args, err := netrules.ParseRuleArguments(rule)
		if err != nil {
			log.Errorf("Error parsing rule %s: %v", rule, err)
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
			log.Warnf("Could not extract chain name or source IP from rule: %s", rule)
			continue
		}

		// Extract container ID from chain name (remove DAYTONA-SB- prefix)
		containerID := strings.TrimPrefix(chainName, "DAYTONA-SB-")
		if containerID == chainName {
			log.Warnf("Invalid chain name format: %s", chainName)
			continue
		}

		// Inspect the container to get its current IP
		container, err := dm.apiClient.ContainerInspect(dm.ctx, containerID)
		if err != nil {
			log.Errorf("Error inspecting container %s: %v", containerID, err)
			// Container doesn't exist, unassign the rules
			if err := dm.netRulesManager.UnassignNetworkRules(containerID); err != nil {
				log.Errorf("Error unassigning rules for non-existent container %s: %v", containerID, err)
			} else {
				log.Infof("Unassigned rules for non-existent container %s", containerID)
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
			log.Warnf("IP mismatch for container %s: rule has %s, container has %s",
				containerID, sourceIP, container.NetworkSettings.IPAddress)

			// Delete only this specific mismatched rule
			if err := dm.netRulesManager.DeleteDockerUserRule(rule); err != nil {
				log.Errorf("Error deleting mismatched rule for container %s: %v", containerID, err)
			} else {
				log.Infof("Deleted mismatched rule for container %s", containerID)
			}
		}
	}
}

// reconcileChains removes orphaned chains for non-existent containers
func (dm *DockerMonitor) reconcileChains() {
	// List all chains that start with DAYTONA-SB-
	chains, err := dm.netRulesManager.ListDaytonaChains()
	if err != nil {
		log.Errorf("Error listing Daytona chains: %v", err)
		return
	}

	for _, chain := range chains {
		// Extract container ID from chain name (remove DAYTONA-SB- prefix)
		containerID := strings.TrimPrefix(chain, "DAYTONA-SB-")
		if containerID == chain {
			log.Warnf("Invalid chain name format: %s", chain)
			continue
		}

		// Check if the container exists
		_, err := dm.apiClient.ContainerInspect(dm.ctx, containerID)
		if err != nil {
			log.Infof("Container %s does not exist, deleting chain %s", containerID, chain)

			// Delete the orphaned chain
			if err := dm.netRulesManager.DeleteChain(chain); err != nil {
				log.Errorf("Error deleting orphaned chain %s: %v", chain, err)
			} else {
				log.Infof("Deleted orphaned chain %s", chain)
			}
		}
	}
}

// reconcilerLoop runs reconciliation every 5 minutes
func (dm *DockerMonitor) reconcilerLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			dm.reconcileNetworkRules()
			dm.reconcileChains()
		}
	}
}
