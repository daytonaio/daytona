// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import "strings"

// SetNetWorkRules creates and configures network rules for a container
func (manager *NetRulesManager) SetNetWorkRules(name string, sourceIp string, networkAllowList string) error {
	// Parse the allowed networks
	allowedNetworks, err := parseCidrNetworks(networkAllowList)
	if err != nil {
		return err
	}

	// Add prefix to chain name
	chainName := formatChainName(name)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Create the chain (ignores if already exists)
	err = manager.ipt.NewChain("filter", chainName)
	if err != nil && !strings.Contains(err.Error(), "Chain already exists") {
		return err
	}

	// Clear existing rules to ensure clean state
	if err := manager.ipt.ClearChain("filter", chainName); err != nil {
		return err
	}

	// Add rules to allow traffic from the specified networks
	for _, network := range allowedNetworks {
		if err := manager.ipt.AppendUnique("filter", chainName, "-j", "RETURN", "-d", network.String(), "-p", "all"); err != nil {
			return err
		}
	}

	// Add a final rule to block all other traffic
	if err := manager.ipt.AppendUnique("filter", chainName, "-j", "DROP", "-p", "all"); err != nil {
		return err
	}

	// Assign the rules to the container (atomic within the same mutex)
	if err := manager.ipt.InsertUnique("filter", "DOCKER-USER", 1, "-j", chainName, "-s", sourceIp, "-p", "all"); err != nil {
		return err
	}

	return manager.saveIptablesRules()
}
