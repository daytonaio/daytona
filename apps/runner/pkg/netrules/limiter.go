// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import "strings"

// SetNetworkLimiter creates and configures network rules for a container
func (manager *NetRulesManager) SetNetworkLimiter(name string, sourceIp string) error {

	// Add prefix to chain name
	chainName := formatChainName(name)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Create the chain (ignores if already exists)
	err := manager.ipt.NewChain("mangle", chainName)
	if err != nil && !strings.Contains(err.Error(), "Chain already exists") {
		return err
	}

	// Clear existing rules to ensure clean state
	if err := manager.ipt.ClearChain("mangle", chainName); err != nil {
		return err
	}

	// Add rule to mark packets
	if err := manager.ipt.AppendUnique("mangle", chainName, "-j", "MARK", "--set-mark", "999"); err != nil {
		return err
	}

	// Assign the rules to the container IP
	if err := manager.ipt.AppendUnique("mangle", "PREROUTING", "-j", chainName, "-s", sourceIp, "-p", "all"); err != nil {
		return err
	}

	return nil
}

// RemoveNetworkLimiter removes the network limiter for a container
func (manager *NetRulesManager) RemoveNetworkLimiter(name string) error {
	chainName := formatChainName(name)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	// First unassign the rules from the container
	rules, err := manager.ipt.List("mangle", "PREROUTING")
	if err != nil {
		return err
	}

	// Find and remove rules that reference our chain
	for _, rule := range rules {
		if strings.Contains(rule, chainName) {
			// Parse the rule to extract arguments
			args, err := ParseRuleArguments(rule)
			if err != nil {
				// Skip malformed rules
				continue
			}

			if err := manager.ipt.Delete("mangle", "PREROUTING", args...); err != nil {
				return err
			}
		}
	}

	// Delete the chain and all its rules
	return manager.ipt.ClearAndDeleteChain("mangle", chainName)
}
