// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

// VMMemoryStats holds memory statistics for a VM collected from the balloon driver
type VMMemoryStats struct {
	DomainID     string // Domain UUID
	DomainName   string // Domain name
	MaxMemoryKiB uint64 // Maximum allowed memory (from domain config)
	ActualKiB    uint64 // Current balloon size (actual allocation)
	UnusedKiB    uint64 // Free memory reported by guest
	UsableKiB    uint64 // Memory guest can use (available - buffers/cache)
	AvailableKiB uint64 // Total memory visible to guest
	RSSKiB       uint64 // Resident set size on host
	LastUpdate   int64  // Timestamp from guest balloon driver (0 = not reporting)
}

// UsedMemoryKiB returns the memory actually being used by the guest
func (s *VMMemoryStats) UsedMemoryKiB() uint64 {
	if s.UnusedKiB > s.ActualKiB {
		return 0
	}
	return s.ActualKiB - s.UnusedKiB
}

// IsBalloonDriverActive returns true if the guest balloon driver is reporting stats
func (s *VMMemoryStats) IsBalloonDriverActive() bool {
	return s.LastUpdate > 0
}

// GetVMMemoryStats retrieves memory statistics for a single VM
func (l *LibVirt) GetVMMemoryStats(ctx context.Context, domainID string) (*VMMemoryStats, error) {
	conn, err := l.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	domain, err := l.LookupDomainBySandboxId(conn, domainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	return l.getMemoryStatsForDomain(domain)
}

// GetAllVMMemoryStats retrieves memory statistics for all running VMs
func (l *LibVirt) GetAllVMMemoryStats(ctx context.Context) (map[string]*VMMemoryStats, error) {
	conn, err := l.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Only get running domains - can't balloon stopped VMs
	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	stats := make(map[string]*VMMemoryStats)

	for _, domain := range domains {
		vmStats, err := l.getMemoryStatsForDomain(&domain)
		if err != nil {
			name, _ := domain.GetName()
			log.Debugf("Failed to get memory stats for domain %s: %v", name, err)
			domain.Free()
			continue
		}

		stats[vmStats.DomainName] = vmStats
		domain.Free()
	}

	return stats, nil
}

// getMemoryStatsForDomain collects memory stats from a domain
func (l *LibVirt) getMemoryStatsForDomain(domain *libvirt.Domain) (*VMMemoryStats, error) {
	name, err := domain.GetName()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain name: %w", err)
	}

	uuid, err := domain.GetUUIDString()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain UUID: %w", err)
	}

	// Get domain info for max memory
	info, err := domain.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info: %w", err)
	}

	stats := &VMMemoryStats{
		DomainID:     uuid,
		DomainName:   name,
		MaxMemoryKiB: info.MaxMem,
		ActualKiB:    info.Memory, // Current memory from domain info
	}

	// Get detailed memory stats from balloon driver
	// Request all available memory stats (0 means all)
	memStats, err := domain.MemoryStats(uint32(libvirt.DOMAIN_MEMORY_STAT_NR), 0)
	if err != nil {
		log.Debugf("Failed to get memory stats for %s (balloon driver may not be active): %v", name, err)
		// Return basic stats without balloon details
		return stats, nil
	}

	// Parse memory stats
	for _, stat := range memStats {
		switch libvirt.DomainMemoryStatTags(stat.Tag) {
		case libvirt.DOMAIN_MEMORY_STAT_ACTUAL_BALLOON:
			stats.ActualKiB = stat.Val
		case libvirt.DOMAIN_MEMORY_STAT_UNUSED:
			stats.UnusedKiB = stat.Val
		case libvirt.DOMAIN_MEMORY_STAT_USABLE:
			stats.UsableKiB = stat.Val
		case libvirt.DOMAIN_MEMORY_STAT_AVAILABLE:
			stats.AvailableKiB = stat.Val
		case libvirt.DOMAIN_MEMORY_STAT_RSS:
			stats.RSSKiB = stat.Val
		case libvirt.DOMAIN_MEMORY_STAT_LAST_UPDATE:
			stats.LastUpdate = int64(stat.Val)
		}
	}

	return stats, nil
}

// SetVMMemory sets the current memory allocation for a VM using the balloon driver
// memoryKiB must be between minMemory and maxMemory for the domain
func (l *LibVirt) SetVMMemory(ctx context.Context, domainID string, memoryKiB uint64) error {
	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	domain, err := l.LookupDomainBySandboxId(conn, domainID)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Get current info to validate
	info, err := domain.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get domain info: %w", err)
	}

	// Validate memory is within bounds
	if memoryKiB > info.MaxMem {
		return fmt.Errorf("requested memory %d KiB exceeds maximum %d KiB", memoryKiB, info.MaxMem)
	}

	// Set memory (this inflates/deflates the balloon)
	// Using DOMAIN_MEM_LIVE flag to affect the running domain
	if err := domain.SetMemoryFlags(memoryKiB, libvirt.DOMAIN_MEM_LIVE); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	name, _ := domain.GetName()
	log.Infof("Set VM %s memory: %d KiB -> %d KiB", name, info.Memory, memoryKiB)

	return nil
}

// SetVMMemoryByName sets memory for a VM by its name
func (l *LibVirt) SetVMMemoryByName(ctx context.Context, domainName string, memoryKiB uint64) error {
	conn, err := l.getConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	domain, err := conn.LookupDomainByName(domainName)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}
	defer domain.Free()

	// Get current info to validate
	info, err := domain.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get domain info: %w", err)
	}

	// Validate memory is within bounds
	if memoryKiB > info.MaxMem {
		return fmt.Errorf("requested memory %d KiB exceeds maximum %d KiB", memoryKiB, info.MaxMem)
	}

	// Set memory (this inflates/deflates the balloon)
	if err := domain.SetMemoryFlags(memoryKiB, libvirt.DOMAIN_MEM_LIVE); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	log.Infof("Set VM %s memory: %d KiB -> %d KiB", domainName, info.Memory, memoryKiB)

	return nil
}
