// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import (
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

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
// The rule is installed just before Docker's terminal `-j RETURN` so that
// per-sandbox egress jumps — which are inserted at DOCKER-USER position 1 —
// still get evaluated first. Otherwise the ACCEPT here would short-circuit
// every per-sandbox egress policy for intra-link traffic.
//
// Idempotent: a no-op if an identical rule already exists at any position.
func (manager *NetRulesManager) AllowSubnetICC(bridgeName, subnet string) error {
	if bridgeName == "" || subnet == "" {
		return nil
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.insertDockerUserBeforeReturn(
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
//
// Installed just before Docker's terminal `-j RETURN` so per-sandbox egress
// jumps (inserted at DOCKER-USER position 1) still fire first — see
// AllowSubnetICC for the rationale.
func (manager *NetRulesManager) AllowBridgeICC(bridgeName string) error {
	if bridgeName == "" {
		return nil
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.insertDockerUserBeforeReturn(
		"-i", bridgeName, "-o", bridgeName, "-j", "ACCEPT")
}

// insertDockerUserBeforeReturn inserts a rule into DOCKER-USER at the position
// just before Docker's terminal `-j RETURN`. This keeps link-network ACCEPT
// rules from short-circuiting per-sandbox egress jumps (which stay at
// position 1). Callers must hold `manager.mu`.
//
// Idempotent: a no-op if an identical rule already exists at any position.
// If no terminal RETURN is present the rule is appended to the chain.
func (manager *NetRulesManager) insertDockerUserBeforeReturn(rulespec ...string) error {
	exists, err := manager.ipt.Exists("filter", "DOCKER-USER", rulespec...)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	pos, hasTerminator, err := dockerUserReturnPos(manager.ipt)
	if err != nil {
		return err
	}
	if !hasTerminator {
		return manager.ipt.Append("filter", "DOCKER-USER", rulespec...)
	}
	return manager.ipt.Insert("filter", "DOCKER-USER", pos, rulespec...)
}

// dockerUserReturnPos returns the 1-indexed position of Docker's default
// `-A DOCKER-USER -j RETURN` terminator. The second return value indicates
// whether the terminator was found. Callers that find it should Insert at
// that position so the new rule lands immediately before the RETURN.
func dockerUserReturnPos(ipt *iptables.IPTables) (int, bool, error) {
	rules, err := ipt.List("filter", "DOCKER-USER")
	if err != nil {
		return 0, false, err
	}

	// ipt.List output is the `iptables -S` rendering: one `-N <chain>` entry
	// followed by one `-A <chain> ...` entry per rule, in chain order.
	pos := 0
	for _, r := range rules {
		if !strings.HasPrefix(r, "-A ") {
			continue
		}
		pos++
		if r == "-A DOCKER-USER -j RETURN" {
			return pos, true, nil
		}
	}
	return 0, false, nil
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
