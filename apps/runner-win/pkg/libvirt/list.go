// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

func (l *LibVirt) DomainList(ctx context.Context, options DomainListOptions) ([]DomainSummary, error) {
	conn, err := l.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	var flags libvirt.ConnectListAllDomainsFlags
	if options.All {
		flags = libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	} else if options.Active {
		flags = libvirt.CONNECT_LIST_DOMAINS_ACTIVE
	} else if options.Inactive {
		flags = libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	} else {
		// Default to all domains
		flags = libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	}

	domains, err := conn.ListAllDomains(flags)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	summaries := make([]DomainSummary, 0, len(domains))
	for _, domain := range domains {
		summary, err := l.getDomainSummary(&domain)
		if err != nil {
			log.Warnf("Failed to get summary for domain: %v", err)
			continue
		}
		summaries = append(summaries, summary)

		// Free the domain
		if err := domain.Free(); err != nil {
			log.Warnf("Failed to free domain: %v", err)
		}
	}

	return summaries, nil
}

func (l *LibVirt) getDomainSummary(domain *libvirt.Domain) (DomainSummary, error) {
	name, err := domain.GetName()
	if err != nil {
		return DomainSummary{}, fmt.Errorf("failed to get domain name: %w", err)
	}

	uuid, err := domain.GetUUIDString()
	if err != nil {
		return DomainSummary{}, fmt.Errorf("failed to get domain UUID: %w", err)
	}

	state, _, err := domain.GetState()
	if err != nil {
		return DomainSummary{}, fmt.Errorf("failed to get domain state: %w", err)
	}

	info, err := domain.GetInfo()
	if err != nil {
		return DomainSummary{}, fmt.Errorf("failed to get domain info: %w", err)
	}

	id, err := domain.GetID()
	if err != nil {
		// Domain might not be running, ID will be -1
		id = 0
	}

	return DomainSummary{
		UUID:   uuid,
		Name:   name,
		State:  DomainState(state),
		Memory: info.Memory,
		VCPUs:  uint(info.NrVirtCpu),
		ID:     int(id),
	}, nil
}

func (l *LibVirt) Info(ctx context.Context) (SystemInfo, error) {
	conn, err := l.getConnection()
	if err != nil {
		return SystemInfo{}, fmt.Errorf("failed to get connection: %w", err)
	}

	hostname, err := conn.GetHostname()
	if err != nil {
		log.Warnf("Failed to get hostname: %v", err)
		hostname = "unknown"
	}

	hvType, err := conn.GetType()
	if err != nil {
		log.Warnf("Failed to get hypervisor type: %v", err)
		hvType = "unknown"
	}

	libVersion, err := conn.GetLibVersion()
	if err != nil {
		log.Warnf("Failed to get libvirt version: %v", err)
		libVersion = 0
	}

	uri, err := conn.GetURI()
	if err != nil {
		log.Warnf("Failed to get URI: %v", err)
		uri = l.libvirtURI
	}

	nodeInfo, err := conn.GetNodeInfo()
	if err != nil {
		log.Warnf("Failed to get node info: %v", err)
	}

	activeDomains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		log.Warnf("Failed to list active domains: %v", err)
	}

	inactiveDomains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		log.Warnf("Failed to list inactive domains: %v", err)
	}

	// Free domains
	for _, domain := range activeDomains {
		_ = domain.Free()
	}
	for _, domain := range inactiveDomains {
		_ = domain.Free()
	}

	totalMemory := uint64(0)
	totalCPUs := 0
	if nodeInfo != nil {
		totalMemory = nodeInfo.Memory
		totalCPUs = int(nodeInfo.Cpus)
	}

	return SystemInfo{
		Hostname:        hostname,
		HypervisorType:  hvType,
		LibvirtVersion:  uint64(libVersion),
		ConnectionURI:   uri,
		TotalMemory:     totalMemory,
		TotalCPUs:       totalCPUs,
		DomainsActive:   len(activeDomains),
		DomainsInactive: len(inactiveDomains),
		DomainsTotal:    len(activeDomains) + len(inactiveDomains),
	}, nil
}

// ContainerList is for compatibility with Docker interface
func (l *LibVirt) ContainerList(ctx context.Context, options DomainListOptions) ([]DomainSummary, error) {
	return l.DomainList(ctx, options)
}
