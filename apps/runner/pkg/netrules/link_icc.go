// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import "strings"

// AllowSubnetICC inserts a DOCKER-USER rule that ACCEPTs traffic that enters
// the host on the named bridge interface and has both source and destination
// IPs inside the given subnet. This is the per-link-network override needed
// when the Docker daemon runs with `icc: false` globally — the per-network
// `com.docker.network.bridge.enable_icc=true` option does not punch through
// the daemon-level ICC block, so without this rule linked sandboxes on the
// same bridge cannot reach each other.
//
// The `-i <bridge>` anchor is required for isolation: without it, the rule
// would also match cross-bridge packets carrying spoofed addresses (a
// privileged container on any other bridge can craft a packet with
// src/dst in the link subnet, and the ACCEPT in DOCKER-USER would
// short-circuit before Docker's inter-bridge isolation drops fire). With the
// input-interface match, packets that don't actually originate on this
// bridge can't satisfy the rule.
//
// The rule is idempotent: InsertUnique is a no-op if an identical rule already
// exists at any position in the chain.
func (manager *NetRulesManager) AllowSubnetICC(bridgeName, subnet string) error {
	if bridgeName == "" || subnet == "" {
		return nil
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.ipt.InsertUnique("filter", "DOCKER-USER", 1,
		"-i", bridgeName, "-s", subnet, "-d", subnet, "-j", "ACCEPT")
}

// RemoveSubnetICC removes the DOCKER-USER ACCEPT rule installed by AllowSubnetICC.
// Missing-rule errors are swallowed so teardown remains idempotent across crashes
// and re-runs.
//
// To stay forward-compatible with runners upgraded across the rule shape change,
// this also attempts to delete the legacy unanchored form (`-s subnet -d subnet`)
// — if a runner installed that variant before the upgrade, teardown should still
// clean it up.
func (manager *NetRulesManager) RemoveSubnetICC(bridgeName, subnet string) error {
	if subnet == "" {
		return nil
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	if bridgeName != "" {
		if err := swallowMissingRule(manager.ipt.Delete("filter", "DOCKER-USER",
			"-i", bridgeName, "-s", subnet, "-d", subnet, "-j", "ACCEPT")); err != nil {
			return err
		}
	}
	return swallowMissingRule(manager.ipt.Delete("filter", "DOCKER-USER",
		"-s", subnet, "-d", subnet, "-j", "ACCEPT"))
}

// AllowBridgeICC inserts a DOCKER-USER rule that ACCEPTs traffic forwarded on a
// single bridge interface in both directions. This mirrors exactly the matcher
// Docker uses to install its own `-j DROP` rule when `icc: false` is in effect
// (`-i <bridge> -o <bridge>`), so the ACCEPT here will short-circuit before
// Docker's DROP can fire regardless of which Docker chain hosts the DROP
// (DOCKER-ISOLATION-STAGE-1 in older versions, DOCKER-FORWARD in newer ones).
//
// We install both this rule and the subnet rule because they catch slightly
// different traffic shapes (interface vs. address) and the cost of the
// duplicate is a single extra iptables rule per link network.
func (manager *NetRulesManager) AllowBridgeICC(bridgeName string) error {
	if bridgeName == "" {
		return nil
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.ipt.InsertUnique("filter", "DOCKER-USER", 1,
		"-i", bridgeName, "-o", bridgeName, "-j", "ACCEPT")
}

// RemoveBridgeICC removes the rule installed by AllowBridgeICC. Missing-rule
// errors are swallowed so teardown is idempotent.
func (manager *NetRulesManager) RemoveBridgeICC(bridgeName string) error {
	if bridgeName == "" {
		return nil
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return swallowMissingRule(manager.ipt.Delete("filter", "DOCKER-USER",
		"-i", bridgeName, "-o", bridgeName, "-j", "ACCEPT"))
}

func swallowMissingRule(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "does a matching rule exist") ||
		strings.Contains(msg, "no chain/target/match by that name") ||
		strings.Contains(msg, "bad rule") {
		return nil
	}
	return err
}
