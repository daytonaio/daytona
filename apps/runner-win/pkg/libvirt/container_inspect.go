// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

// IP address wait configuration
const (
	ipAddressWaitTimeout  = 60 * time.Second       // Max time to wait for IP address
	ipAddressWaitInterval = 500 * time.Millisecond // Faster polling interval (500ms)
)

func (l *LibVirt) ContainerInspect(ctx context.Context, domainId string) (DomainInfo, error) {
	conn, err := l.getConnection()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain by UUID first, then by name
	domain, err := conn.LookupDomainByUUIDString(domainId)
	if err != nil {
		// Try by name
		domain, err = conn.LookupDomainByName(domainId)
		if err != nil {
			return DomainInfo{}, fmt.Errorf("domain not found: %w", err)
		}
	}
	defer domain.Free()

	uuid, err := domain.GetUUIDString()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get UUID: %w", err)
	}

	name, err := domain.GetName()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get name: %w", err)
	}

	state, _, err := domain.GetState()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get state: %w", err)
	}

	info, err := domain.GetInfo()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get info: %w", err)
	}

	id, err := domain.GetID()
	if err != nil {
		// GetID only works for running domains, not paused/stopped ones - this is expected
		log.Debugf("Failed to get domain ID (domain may not be running): %v", err)
		id = 0
	}

	// Get IP address with retry logic (VM needs time to boot and get DHCP lease)
	ipAddress := l.waitForDomainIP(ctx, conn, domain, name)

	return DomainInfo{
		UUID:      uuid,
		Name:      name,
		State:     DomainState(state),
		Memory:    info.Memory,
		MaxMemory: info.MaxMem,
		VCPUs:     uint(info.NrVirtCpu),
		CPUTime:   info.CpuTime,
		ID:        int(id),
		Metadata:  make(map[string]string),
		IPAddress: ipAddress,
	}, nil
}

// waitForDomainIP waits for the domain to get an IP address with retry logic
func (l *LibVirt) waitForDomainIP(ctx context.Context, conn *libvirt.Connect, domain *libvirt.Domain, domainName string) string {
	// Always get actual IP from DHCP lease, not pre-calculated reservation
	// The DHCP reservation may not work reliably with Windows VMs
	if ip := l.getDomainIP(conn, domain); ip != "" {
		log.Infof("Domain %s has actual IP: %s", domainName, ip)
		return ip
	}

	log.Infof("Waiting for domain %s to get an IP address...", domainName)

	deadline := time.Now().Add(ipAddressWaitTimeout)
	ticker := time.NewTicker(ipAddressWaitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warnf("Context cancelled while waiting for IP address for domain %s", domainName)
			return ""
		case <-ticker.C:
			if time.Now().After(deadline) {
				log.Warnf("Timeout waiting for IP address for domain %s after %v", domainName, ipAddressWaitTimeout)
				return ""
			}

			if ip := l.getDomainIP(conn, domain); ip != "" {
				log.Infof("Domain %s got IP address: %s", domainName, ip)
				return ip
			}

			log.Debugf("Still waiting for IP address for domain %s...", domainName)
		}
	}
}

// getDomainIP gets IP address by querying network DHCP leases using MAC address
func (l *LibVirt) getDomainIP(conn *libvirt.Connect, domain *libvirt.Domain) string {
	// Get domain XML to extract MAC address and network name
	xmlDesc, err := domain.GetXMLDesc(0)
	if err != nil {
		log.Debugf("Failed to get domain XML: %v", err)
		return ""
	}

	// Parse domain XML
	var domainXML libvirtxml.Domain
	if err := domainXML.Unmarshal(xmlDesc); err != nil {
		log.Debugf("Failed to parse domain XML: %v", err)
		return ""
	}

	// Find network interfaces and their MAC addresses
	if domainXML.Devices == nil {
		return ""
	}

	for _, iface := range domainXML.Devices.Interfaces {
		var networkName string
		var macAddress string

		// Get MAC address
		if iface.MAC != nil {
			macAddress = iface.MAC.Address
		}
		if macAddress == "" {
			continue
		}

		// Get network name from source
		if iface.Source != nil && iface.Source.Network != nil {
			networkName = iface.Source.Network.Network
		}
		if networkName == "" {
			networkName = "default" // Fall back to default network
		}

		// Get IP from network DHCP leases
		ip := l.getIPFromNetworkLease(conn, networkName, macAddress)
		if ip != "" {
			return ip
		}
	}

	return ""
}

// getIPFromNetworkLease queries a network's DHCP leases for a specific MAC address
func (l *LibVirt) getIPFromNetworkLease(conn *libvirt.Connect, networkName, macAddress string) string {
	network, err := conn.LookupNetworkByName(networkName)
	if err != nil {
		log.Debugf("Failed to lookup network %s: %v", networkName, err)
		return ""
	}
	defer network.Free()

	// Get DHCP leases
	leases, err := network.GetDHCPLeases()
	if err != nil {
		log.Debugf("Failed to get DHCP leases for network %s: %v", networkName, err)
		return ""
	}

	for _, lease := range leases {
		if lease.Mac == macAddress {
			return lease.IPaddr
		}
	}

	return ""
}

// ContainerInspectBasic returns domain info WITHOUT waiting for IP address
// Use this for metrics collection where IP is not needed
func (l *LibVirt) ContainerInspectBasic(ctx context.Context, domainId string) (DomainInfo, error) {
	conn, err := l.getConnection()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get connection: %w", err)
	}

	// Try to find domain by UUID first, then by name
	domain, err := conn.LookupDomainByUUIDString(domainId)
	if err != nil {
		// Try by name
		domain, err = conn.LookupDomainByName(domainId)
		if err != nil {
			return DomainInfo{}, fmt.Errorf("domain not found: %w", err)
		}
	}
	defer domain.Free()

	uuid, err := domain.GetUUIDString()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get UUID: %w", err)
	}

	name, err := domain.GetName()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get name: %w", err)
	}

	state, _, err := domain.GetState()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get state: %w", err)
	}

	info, err := domain.GetInfo()
	if err != nil {
		return DomainInfo{}, fmt.Errorf("failed to get info: %w", err)
	}

	id, err := domain.GetID()
	if err != nil {
		// GetID only works for running domains, not paused/stopped ones - this is expected
		log.Debugf("Failed to get domain ID (domain may not be running): %v", err)
		id = 0
	}

	// Return info WITHOUT waiting for IP address
	return DomainInfo{
		UUID:      uuid,
		Name:      name,
		State:     DomainState(state),
		Memory:    info.Memory,
		MaxMemory: info.MaxMem,
		VCPUs:     uint(info.NrVirtCpu),
		CPUTime:   info.CpuTime,
		ID:        int(id),
		Metadata:  make(map[string]string),
	}, nil
}
