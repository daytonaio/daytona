// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os/exec"
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

	// Create the disk for the VM
	diskPath, err := l.createDisk(ctx, sandboxDto)
	if err != nil {
		return "", "", fmt.Errorf("failed to create disk: %w", err)
	}

	// Build domain XML configuration
	domainXML, err := l.buildDomainXML(sandboxDto, diskPath)
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

	// Start the domain automatically after creation
	log.Infof("Starting domain %s after creation", name)
	if err := domain.Create(); err != nil {
		return "", "", fmt.Errorf("failed to start domain: %w", err)
	}

	// Wait for domain to be running
	if err := l.waitForDomainRunning(ctx, domain); err != nil {
		log.Warnf("Domain created but failed to start properly: %v", err)
		if l.statesCache != nil {
			l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateStopped)
		}
		return uuid, name, fmt.Errorf("domain created but failed to start: %w", err)
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateStarted)
	}
	log.Infof("Domain %s created and started successfully with UUID %s", name, uuid)

	return uuid, name, nil
}

// Minimum resource requirements for Windows VMs
const (
	minWindowsMemoryMB = 4096 // 4 GB minimum for Windows 11
	minWindowsVCPUs    = 2    // 2 vCPUs minimum for Windows 11
)

func (l *LibVirt) buildDomainXML(sandboxDto dto.CreateSandboxDTO, diskPath string) (string, error) {
	// Get memory in MB, enforce minimum for Windows
	memoryMB := sandboxDto.MemoryQuota
	if memoryMB < minWindowsMemoryMB {
		log.Warnf("Memory quota %d MB is below minimum %d MB for Windows, using minimum", memoryMB, minWindowsMemoryMB)
		memoryMB = minWindowsMemoryMB
	}
	// Convert memory from MB to KiB (libvirt uses KiB)
	memoryKiB := uint(memoryMB * 1024)

	// Use CPU quota as number of VCPUs, enforce minimum for Windows
	vcpus := uint(sandboxDto.CpuQuota)
	if vcpus < minWindowsVCPUs {
		log.Warnf("CPU quota %d is below minimum %d for Windows, using minimum", vcpus, minWindowsVCPUs)
		vcpus = minWindowsVCPUs
	}

	// Build the domain configuration for Windows 11 (requires UEFI + SecureBoot)
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
			Firmware: "efi",
			Type: &libvirtxml.DomainOSType{
				Type:    "hvm",
				Arch:    "x86_64",
				Machine: "q35",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{Dev: "hd"},
			},
			Loader: &libvirtxml.DomainLoader{
				Readonly: "yes",
				Secure:   "yes",
				Type:     "pflash",
				Path:     "/usr/share/OVMF/OVMF_CODE_4M.ms.fd",
			},
			NVRam: &libvirtxml.DomainNVRam{
				Template: "/usr/share/OVMF/OVMF_VARS_4M.ms.fd",
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
			HyperV: &libvirtxml.DomainFeatureHyperV{
				Mode:    "custom",
				Relaxed: &libvirtxml.DomainFeatureState{State: "on"},
				VAPIC:   &libvirtxml.DomainFeatureState{State: "on"},
				Spinlocks: &libvirtxml.DomainFeatureHyperVSpinlocks{
					DomainFeatureState: libvirtxml.DomainFeatureState{State: "on"},
					Retries:            8191,
				},
			},
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
							File: diskPath,
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

func (l *LibVirt) createDisk(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	basePath := "/var/lib/libvirt/images"

	// Use a default base image for now
	// Snapshots are not used in this implementation yet
	baseImage := "win11-clone-base"

	baseImagePath := filepath.Join(basePath, fmt.Sprintf("%s.qcow2", baseImage))
	newDiskPath := filepath.Join(basePath, fmt.Sprintf("%s.qcow2", sandboxDto.Id))

	log.Infof("Creating disk %s from base image %s", newDiskPath, baseImagePath)

	// Create a qcow2 overlay disk with the base image as backing file
	// This uses copy-on-write, so each VM has its own changes without modifying the base
	createCmd := fmt.Sprintf("qemu-img create -f qcow2 -F qcow2 -b %s %s", baseImagePath, newDiskPath)

	log.Infof("Executing on remote server: %s", createCmd)

	// Execute the command on the remote server via SSH
	// We extract the hostname from the libvirt URI
	// For qemu+ssh://root@h1001.blinkbox.dev/system, we need h1001.blinkbox.dev
	host := l.extractHostFromURI()
	if host == "" {
		return "", fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	// Execute the disk creation command via SSH
	cmd := exec.CommandContext(ctx, "ssh", host, createCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to create disk: %v, output: %s", err, string(output))
		return "", fmt.Errorf("failed to create disk on remote server: %w (output: %s)", err, string(output))
	}

	log.Infof("Successfully created disk %s", newDiskPath)
	return newDiskPath, nil
}

func (l *LibVirt) extractHostFromURI() string {
	// Extract host from URI like "qemu+ssh://root@h1001.blinkbox.dev/system"
	uri := l.libvirtURI
	if len(uri) == 0 {
		return ""
	}

	// Find the host part between @ and /
	atIndex := -1
	for i := 0; i < len(uri); i++ {
		if uri[i] == '@' {
			atIndex = i
			break
		}
	}
	if atIndex == -1 {
		return ""
	}

	slashIndex := -1
	for i := atIndex; i < len(uri); i++ {
		if uri[i] == '/' {
			slashIndex = i
			break
		}
	}
	if slashIndex == -1 {
		return uri[atIndex+1:]
	}

	return uri[atIndex+1 : slashIndex]
}
