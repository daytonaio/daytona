// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/coreos/go-iptables/iptables"
)

// NetRulesManager provides thread-safe operations for managing network rules
type NetRulesManager struct {
	ipt        *iptables.IPTables
	mu         sync.Mutex
	persistent bool
}

// NewNetRulesManager creates a new instance of NetRulesManager
func NewNetRulesManager(persistent bool) (*NetRulesManager, error) {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}

	return &NetRulesManager{
		ipt:        ipt,
		persistent: persistent,
	}, nil
}

// saveIptablesRules saves the current iptables rules to make them persistent
func (manager *NetRulesManager) saveIptablesRules() error {
	if manager.persistent {
		cmd := exec.Command("sh", "-c", "iptables-save > /etc/iptables/rules.v4")
		return cmd.Run()
	}
	return nil
}

// ListDaytonaRules returns all DOCKER-USER rules that jump to Daytona chains
func (manager *NetRulesManager) ListDaytonaRules() ([]string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	rules, err := manager.ipt.List("filter", "DOCKER-USER")
	if err != nil {
		return nil, err
	}

	var daytonaRules []string
	for _, rule := range rules {
		if strings.Contains(rule, ChainPrefix) {
			daytonaRules = append(daytonaRules, rule)
		}
	}

	return daytonaRules, nil
}

// DeleteDockerUserRule deletes a specific rule from DOCKER-USER chain
func (manager *NetRulesManager) DeleteDockerUserRule(rule string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	args, err := ParseRuleArguments(rule)
	if err != nil {
		return err
	}

	if err := manager.ipt.Delete("filter", "DOCKER-USER", args...); err != nil {
		return err
	}

	return manager.saveIptablesRules()
}

// ListDaytonaChains returns all chains that start with DAYTONA-SB-
func (manager *NetRulesManager) ListDaytonaChains() ([]string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	chains, err := manager.ipt.ListChains("filter")
	if err != nil {
		return nil, err
	}

	var daytonaChains []string
	for _, chain := range chains {
		if strings.HasPrefix(chain, ChainPrefix) {
			daytonaChains = append(daytonaChains, chain)
		}
	}

	return daytonaChains, nil
}

// DeleteChain deletes a specific chain
func (manager *NetRulesManager) DeleteChain(chainName string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if err := manager.ipt.ClearAndDeleteChain("filter", chainName); err != nil {
		return err
	}

	return manager.saveIptablesRules()
}
