// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
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
		log.Warnf("Failed to get domain ID: %v", err)
		id = 0
	}

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
