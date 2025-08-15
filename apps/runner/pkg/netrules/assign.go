// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

// AssignNetworkRules assigns network rules to a container by inserting a rule in DOCKER-USER chain
func (manager *NetRulesManager) AssignNetworkRules(name string, sourceIp string) error {
	// Add prefix to chain name
	chainName := formatChainName(name)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	if err := manager.ipt.InsertUnique("filter", "DOCKER-USER", 1, "-j", chainName, "-s", sourceIp, "-p", "all"); err != nil {
		return err
	}

	return manager.saveIptablesRules()
}
