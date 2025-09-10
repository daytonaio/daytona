// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import "strings"

// UnassignNetworkRules removes network rules assignment from a container
func (manager *NetRulesManager) UnassignNetworkRules(name string) error {
	// Add prefix to chain name
	chainName := formatChainName(name)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Get all rules from DOCKER-USER chain
	rules, err := manager.ipt.List("filter", "DOCKER-USER")
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

			if err := manager.ipt.Delete("filter", "DOCKER-USER", args...); err != nil {
				return err
			}
		}
	}

	return manager.saveIptablesRules()
}
