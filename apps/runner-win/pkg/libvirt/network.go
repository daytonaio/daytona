// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Network configuration constants
const (
	NetworkName = "default"

	// IP range for sandboxes: 10.100.1.0 - 10.100.15.254
	// This gives us 3838 usable IPs (15*256 - 2 for network/broadcast adjustments)
	BaseIPFirstOctet  = 10
	BaseIPSecondOctet = 100
	BaseIPThirdOctet  = 1  // Start from 10.100.1.x
	MaxThirdOctet     = 15 // End at 10.100.15.x
	MaxSandboxes      = 3838
)

// GenerateMACFromSandboxID creates a deterministic MAC address from sandbox ID
// Format: 52:54:00:xx:xx:xx (QEMU/KVM locally administered prefix)
func GenerateMACFromSandboxID(sandboxId string) string {
	hash := sha256.Sum256([]byte(sandboxId))

	// Use QEMU/KVM prefix 52:54:00 and last 3 bytes from hash
	mac := fmt.Sprintf("52:54:00:%02x:%02x:%02x", hash[0], hash[1], hash[2])
	return mac
}

// CalculateIPFromSandboxID derives a deterministic IP from sandbox ID
// Returns IP in range 10.100.1.0 - 10.100.15.254
func CalculateIPFromSandboxID(sandboxId string) string {
	hash := sha256.Sum256([]byte(sandboxId))

	// Use first 4 bytes of hash to get offset
	offset := binary.BigEndian.Uint32(hash[:4]) % MaxSandboxes

	// Calculate third and fourth octets
	// offset 0    -> 10.100.1.0
	// offset 255  -> 10.100.1.255
	// offset 256  -> 10.100.2.0
	// offset 3837 -> 10.100.15.253
	thirdOctet := BaseIPThirdOctet + int(offset/256)
	fourthOctet := int(offset % 256)

	// Ensure we don't exceed the range
	if thirdOctet > MaxThirdOctet {
		thirdOctet = MaxThirdOctet
		fourthOctet = 254
	}

	ip := fmt.Sprintf("%d.%d.%d.%d", BaseIPFirstOctet, BaseIPSecondOctet, thirdOctet, fourthOctet)
	return ip
}

// AddDHCPReservation adds a MAC->IP mapping to the network's DHCP configuration
// This is done via virsh net-update command
func (l *LibVirt) AddDHCPReservation(mac, ip, hostname string) error {
	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	network, err := conn.LookupNetworkByName(NetworkName)
	if err != nil {
		return fmt.Errorf("failed to lookup network %s: %w", NetworkName, err)
	}
	defer network.Free()

	// Build the DHCP host XML
	hostXML := fmt.Sprintf(`<host mac="%s" ip="%s" name="%s"/>`, mac, ip, hostname)

	log.Infof("Adding DHCP reservation: MAC=%s IP=%s hostname=%s", mac, ip, hostname)

	// Add the DHCP host entry
	// Flags: VIR_NETWORK_UPDATE_AFFECT_LIVE | VIR_NETWORK_UPDATE_AFFECT_CONFIG = 3
	err = network.Update(
		2,  // VIR_NETWORK_UPDATE_COMMAND_ADD_LAST
		4,  // VIR_NETWORK_SECTION_IP_DHCP_HOST
		-1, // parentIndex (-1 for auto)
		hostXML,
		3, // VIR_NETWORK_UPDATE_AFFECT_LIVE | VIR_NETWORK_UPDATE_AFFECT_CONFIG
	)
	if err != nil {
		// Check if it's a duplicate error (reservation already exists)
		log.Warnf("Failed to add DHCP reservation (may already exist): %v", err)
		// Try to delete and re-add
		_ = l.RemoveDHCPReservation(mac)
		err = network.Update(2, 4, -1, hostXML, 3)
		if err != nil {
			return fmt.Errorf("failed to add DHCP reservation: %w", err)
		}
	}

	log.Infof("DHCP reservation added successfully")
	return nil
}

// RemoveDHCPReservation removes a MAC->IP mapping from the network's DHCP configuration
func (l *LibVirt) RemoveDHCPReservation(mac string) error {
	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	network, err := conn.LookupNetworkByName(NetworkName)
	if err != nil {
		return fmt.Errorf("failed to lookup network %s: %w", NetworkName, err)
	}
	defer network.Free()

	// Build the DHCP host XML (only MAC is needed for deletion)
	hostXML := fmt.Sprintf(`<host mac="%s"/>`, mac)

	log.Infof("Removing DHCP reservation for MAC=%s", mac)

	// Delete the DHCP host entry
	// Flags: VIR_NETWORK_UPDATE_AFFECT_LIVE | VIR_NETWORK_UPDATE_AFFECT_CONFIG = 3
	err = network.Update(
		3,  // VIR_NETWORK_UPDATE_COMMAND_DELETE
		4,  // VIR_NETWORK_SECTION_IP_DHCP_HOST
		-1, // parentIndex (-1 for auto)
		hostXML,
		3, // VIR_NETWORK_UPDATE_AFFECT_LIVE | VIR_NETWORK_UPDATE_AFFECT_CONFIG
	)
	if err != nil {
		log.Warnf("Failed to remove DHCP reservation (may not exist): %v", err)
		// Don't return error - it's okay if reservation doesn't exist
		return nil
	}

	log.Infof("DHCP reservation removed successfully")
	return nil
}

// GetReservedIP returns the pre-calculated IP for a sandbox ID
// This can be called immediately without waiting for DHCP
func GetReservedIP(sandboxId string) string {
	return CalculateIPFromSandboxID(sandboxId)
}

// GetReservedMAC returns the pre-calculated MAC for a sandbox ID
func GetReservedMAC(sandboxId string) string {
	return GenerateMACFromSandboxID(sandboxId)
}

// UpdateNetworkSettings updates network settings for a domain
// This is a stub implementation for now - actual network firewall rules
// would need to be implemented using iptables or nftables
func (l *LibVirt) UpdateNetworkSettings(ctx context.Context, domainId string, settings interface{}) error {
	log.Warnf("UpdateNetworkSettings not fully implemented for libvirt domain %s", domainId)
	// TODO: Implement network blocking/allow list using iptables/nftables
	// This would require:
	// 1. Getting the domain's IP address
	// 2. Creating firewall rules to block/allow traffic
	return nil
}
