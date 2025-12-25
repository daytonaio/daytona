// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package netrules

import (
	"fmt"
	"net"
	"strings"
)

const (
	// ChainPrefix is the prefix used for all Daytona sandbox chains
	ChainPrefix = "DAYTONA-SB-"
)

// ParseCidrNetworks parses a comma-separated list of CIDR networks and returns them as an array
func parseCidrNetworks(networks string) ([]*net.IPNet, error) {
	networkList := strings.Split(networks, ",")
	var cidrs []*net.IPNet

	for _, network := range networkList {
		trimmedNetwork := strings.TrimSpace(network)
		if trimmedNetwork == "" {
			continue
		}

		_, ipNet, err := net.ParseCIDR(trimmedNetwork)
		if err != nil {
			return nil, err
		}
		cidrs = append(cidrs, ipNet)
	}

	return cidrs, nil
}

// ParseRuleArguments parses an iptables rule string and returns the arguments
func ParseRuleArguments(rule string) ([]string, error) {
	// Remove the "-A CHAIN_NAME " prefix and split into arguments
	// Rule format: "-A DOCKER-USER -s 172.17.0.2/32 -j chain_name"
	if strings.HasPrefix(rule, "-A ") {
		// Find the first space after "-A CHAIN_NAME"
		parts := strings.SplitN(rule, " ", 3)
		if len(parts) >= 3 {
			return strings.Fields(parts[2]), nil
		}
	}
	return nil, fmt.Errorf("invalid rule format: %s", rule)
}

// formatChainName adds the DAYTONA-SB- prefix to a chain name if it doesn't already have it
func formatChainName(name string) string {
	if strings.HasPrefix(name, ChainPrefix) {
		return name
	}
	return ChainPrefix + name
}
