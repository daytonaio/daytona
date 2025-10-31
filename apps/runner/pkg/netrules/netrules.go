// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

// NetRulesManager provides thread-safe operations for managing network rules
type NetRulesManager struct {
	ipt        *iptables.IPTables
	mu         sync.Mutex
	persistent bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewNetRulesManager creates a new instance of NetRulesManager
func NewNetRulesManager(persistent bool) (*NetRulesManager, error) {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &NetRulesManager{
		ipt:        ipt,
		persistent: persistent,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

func (manager *NetRulesManager) Start() error {
	// Start periodic reconciliation
	if manager.persistent {
		go manager.persistRulesLoop()
	}

	return nil
}

// Stop gracefully stops the NetRulesManager
func (manager *NetRulesManager) Stop() {
	if manager.cancel != nil {
		manager.cancel()
	}
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
func (manager *NetRulesManager) ListDaytonaRules(table string, chain string) ([]string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	rules, err := manager.ipt.List(table, chain)
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

// DeleteChainRule deletes a specific rule from a specific chain
func (manager *NetRulesManager) DeleteChainRule(table string, chain string, rule string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	args, err := ParseRuleArguments(rule)
	if err != nil {
		return err
	}

	return manager.ipt.Delete(table, chain, args...)
}

// ListDaytonaChains returns all chains that start with DAYTONA-SB-
func (manager *NetRulesManager) ListDaytonaChains(table string) ([]string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	chains, err := manager.ipt.ListChains(table)
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

// ClearAndDeleteChain deletes a specific table chain
func (manager *NetRulesManager) ClearAndDeleteChain(table string, name string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.ipt.ClearAndDeleteChain(table, name)
}

// persistRulesLoop persists the iptables rules
func (manager *NetRulesManager) persistRulesLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	slog.Info("Starting iptables persistence loop")

	for {
		select {
		case <-manager.ctx.Done():
			slog.Info("Stopping iptables persistence loop")
			return
		case <-ticker.C:
			if err := manager.saveIptablesRules(); err != nil {
				slog.Error("Failed to save iptables rules", "error", err)
			}
		}
	}
}
