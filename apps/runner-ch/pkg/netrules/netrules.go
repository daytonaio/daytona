// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import (
	"fmt"
	"sync"

	"github.com/coreos/go-iptables/iptables"
	log "github.com/sirupsen/logrus"
)

// NetRulesManager manages network rules for sandboxes
type NetRulesManager struct {
	ipt           *iptables.IPTables
	sandboxRules  map[string][]string // sandboxId -> list of rules
	mutex         sync.Mutex
	bridgeNetwork string
}

// NewNetRulesManager creates a new network rules manager
func NewNetRulesManager(bridgeNetwork string) (*NetRulesManager, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize iptables: %w", err)
	}

	if bridgeNetwork == "" {
		bridgeNetwork = "10.0.0.0/24"
	}

	return &NetRulesManager{
		ipt:           ipt,
		sandboxRules:  make(map[string][]string),
		bridgeNetwork: bridgeNetwork,
	}, nil
}

// SetNetworkBlockAll blocks all outgoing traffic for a sandbox
func (n *NetRulesManager) SetNetworkBlockAll(sandboxId string, ip string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Block all outgoing traffic from the sandbox IP
	rule := []string{"-s", ip, "-j", "DROP"}

	if err := n.ipt.AppendUnique("filter", "FORWARD", rule...); err != nil {
		return fmt.Errorf("failed to add block rule: %w", err)
	}

	n.sandboxRules[sandboxId] = append(n.sandboxRules[sandboxId], fmt.Sprintf("-s %s -j DROP", ip))

	log.Infof("Blocked all network traffic for sandbox %s (IP: %s)", sandboxId, ip)
	return nil
}

// SetNetworkAllowList sets an allow list for a sandbox (blocking all other traffic)
func (n *NetRulesManager) SetNetworkAllowList(sandboxId string, ip string, allowList string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// First allow traffic to the allow list
	allowRule := []string{"-s", ip, "-d", allowList, "-j", "ACCEPT"}
	if err := n.ipt.InsertUnique("filter", "FORWARD", 1, allowRule...); err != nil {
		return fmt.Errorf("failed to add allow rule: %w", err)
	}

	// Then block all other traffic
	blockRule := []string{"-s", ip, "-j", "DROP"}
	if err := n.ipt.AppendUnique("filter", "FORWARD", blockRule...); err != nil {
		return fmt.Errorf("failed to add block rule: %w", err)
	}

	n.sandboxRules[sandboxId] = append(n.sandboxRules[sandboxId],
		fmt.Sprintf("-s %s -d %s -j ACCEPT", ip, allowList),
		fmt.Sprintf("-s %s -j DROP", ip))

	log.Infof("Set network allow list for sandbox %s (IP: %s, allow: %s)", sandboxId, ip, allowList)
	return nil
}

// ClearNetworkRules clears all network rules for a sandbox
func (n *NetRulesManager) ClearNetworkRules(sandboxId string, ip string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Remove any block rules
	_ = n.ipt.Delete("filter", "FORWARD", "-s", ip, "-j", "DROP")

	// Clear tracked rules for this sandbox
	delete(n.sandboxRules, sandboxId)

	log.Infof("Cleared network rules for sandbox %s", sandboxId)
	return nil
}

// UpdateNetworkSettings updates network settings for a sandbox
func (n *NetRulesManager) UpdateNetworkSettings(sandboxId string, ip string, blockAll *bool, allowList *string) error {
	// First clear existing rules
	if err := n.ClearNetworkRules(sandboxId, ip); err != nil {
		log.Warnf("Failed to clear existing rules: %v", err)
	}

	// Apply new rules
	if blockAll != nil && *blockAll {
		if allowList != nil && *allowList != "" {
			return n.SetNetworkAllowList(sandboxId, ip, *allowList)
		}
		return n.SetNetworkBlockAll(sandboxId, ip)
	}

	return nil
}
