// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirtxml"
)

func (l *LibVirt) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, string, error) {
	domainMutex := l.getDomainMutex(sandboxDto.Id)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)
	}

	conn, err := l.getConnection()
	if err != nil {
		return "", "", fmt.Errorf("failed to get connection: %w", err)
	}

	// Build domain XML configuration
	domainXML, err := l.buildDomainXML(sandboxDto)
	if err != nil {
		return "", "", fmt.Errorf("failed to build domain XML: %w", err)
	}

	// Define the domain
	log.Infof("Defining domain %s", sandboxDto.Id)
	domain, err := conn.DomainDefineXML(domainXML)
	if err != nil {
		return "", "", fmt.Errorf("failed to define domain: %w", err)
	}
	defer domain.Free()

	uuid, err := domain.GetUUIDString()
	if err != nil {
		return "", "", fmt.Errorf("failed to get domain UUID: %w", err)
	}

	name, err := domain.GetName()
	if err != nil {
		return "", "", fmt.Errorf("failed to get domain name: %w", err)
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateStopped)
	}
	log.Infof("Domain %s created successfully with UUID %s", name, uuid)

	return uuid, name, nil
}

func (l *LibVirt) buildDomainXML(sandboxDto dto.CreateSandboxDTO) (string, error) {
	// Convert memory from MB to KiB (libvirt uses KiB)
	memoryKiB := uint(sandboxDto.MemoryQuota * 1024)

	// Use CPU quota as number of VCPUs (simplified for first implementation)
	vcpus := uint(sandboxDto.CpuQuota)
	if vcpus == 0 {
		vcpus = 1
	}

	// Build the domain configuration
	domainCfg := &libvirtxml.Domain{
		Type: "kvm",
		Name: sandboxDto.Id,
		Memory: &libvirtxml.DomainMemory{
			Value: memoryKiB,
			Unit:  "KiB",
		},
		CurrentMemory: &libvirtxml.DomainCurrentMemory{
			Value: memoryKiB,
			Unit:  "KiB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: vcpus,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
				Arch: "x86_64",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{Dev: "hd"},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-passthrough",
		},
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: l.getDiskPath(sandboxDto),
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: "default",
						},
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
			Graphics: []libvirtxml.DomainGraphic{
				{
					VNC: &libvirtxml.DomainGraphicVNC{
						Port:        -1, // Auto-allocate
						AutoPort:    "yes",
						Listen:      "0.0.0.0",
						SharePolicy: "allow-exclusive",
					},
				},
			},
			Consoles: []libvirtxml.DomainConsole{
				{
					Source: &libvirtxml.DomainChardevSource{
						Pty: &libvirtxml.DomainChardevSourcePty{},
					},
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "serial",
						Port: new(uint),
					},
				},
			},
		},
	}

	// Marshal to XML
	xml, err := domainCfg.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal domain XML: %w", err)
	}

	log.Debugf("Generated domain XML:\n%s", xml)
	return xml, nil
}

func (l *LibVirt) getDiskPath(sandboxDto dto.CreateSandboxDTO) string {
	// For now, use a simple path based on the snapshot name
	// In a real implementation, this would clone from the base image
	// and create a new disk in a proper location
	basePath := "/var/lib/libvirt/images"

	// Use the snapshot as the base image name
	baseImage := sandboxDto.Snapshot
	if baseImage == "" {
		baseImage = "default"
	}

	// Create a disk name based on sandbox ID
	diskName := fmt.Sprintf("%s.qcow2", sandboxDto.Id)

	return filepath.Join(basePath, diskName)
}
