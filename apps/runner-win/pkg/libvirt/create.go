// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirtxml"
)

func (l *LibVirt) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, string, error) {
	domainMutex := l.getDomainMutex(sandboxDto.Id)
	domainMutex.Lock()
	defer domainMutex.Unlock()

	conn, err := l.getConnection()
	if err != nil {
		return "", "", fmt.Errorf("failed to get connection: %w", err)
	}

	// Check if domain already exists (idempotency)
	existingDomain, err := conn.LookupDomainByName(sandboxDto.Id)
	if err == nil && existingDomain != nil {
		// Domain already exists - return success
		uuid, _ := existingDomain.GetUUIDString()
		name, _ := existingDomain.GetName()
		log.Infof("Domain %s already exists (UUID: %s), returning existing", sandboxDto.Id, uuid)
		existingDomain.Free()
		if l.statesCache != nil {
			l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateStarted)
		}
		return uuid, name, nil
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)
	}

	// Generate deterministic MAC and IP from sandbox ID
	mac := GenerateMACFromSandboxID(sandboxDto.Id)
	ip := CalculateIPFromSandboxID(sandboxDto.Id)
	log.Infof("Generated MAC=%s IP=%s for sandbox %s", mac, ip, sandboxDto.Id)

	// Add DHCP reservation BEFORE starting the VM
	// This ensures the VM gets the expected IP immediately
	if err := l.AddDHCPReservation(mac, ip, sandboxDto.Id); err != nil {
		log.Warnf("Failed to add DHCP reservation: %v (continuing anyway)", err)
	}

	// Create the disk for the VM (linked clone from base image)
	diskPath, err := l.createDisk(ctx, sandboxDto)
	if err != nil {
		return "", "", fmt.Errorf("failed to create disk: %w", err)
	}

	// Copy NVRAM for UEFI boot
	nvramPath, err := l.copyNVRAM(ctx, sandboxDto.Id)
	if err != nil {
		return "", "", fmt.Errorf("failed to copy NVRAM: %w", err)
	}

	// Build domain XML configuration with specific MAC address and NVRAM path
	domainXML, err := l.buildDomainXML(sandboxDto, diskPath, nvramPath, mac)
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

	// Wait for daemon API to be ready
	if err := l.waitForDaemonReady(ctx, sandboxDto.Id, ip); err != nil {
		log.Warnf("Domain started but daemon not ready: %v", err)
		// Don't fail - the sandbox is running, daemon might just be slow
	}

	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateStarted)
	}
	log.Infof("Domain %s created and started successfully with UUID %s, reserved IP %s", name, uuid, ip)

	return uuid, name, nil
}

// Minimum resource requirements for Windows VMs
const (
	minWindowsMemoryMB = 4096 // 4 GB minimum for Windows 11
	minWindowsVCPUs    = 2    // 2 vCPUs minimum for Windows 11
)

func (l *LibVirt) buildDomainXML(sandboxDto dto.CreateSandboxDTO, diskPath string, nvramPath string, macAddress string) (string, error) {
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
				NVRam:    nvramPath,
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
					MAC: &libvirtxml.DomainInterfaceMAC{
						Address: macAddress,
					},
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

// Base image and NVRAM paths for Windows sandboxes
const (
	baseImagePath  = "/var/lib/libvirt/images/winserver-desktop-base.qcow2"
	templateNVRAM  = "/var/lib/libvirt/qemu/nvram/winserver-desktop-base_VARS.fd"
	imagesBasePath = "/var/lib/libvirt/images"
	nvramBasePath  = "/var/lib/libvirt/qemu/nvram"
)

func (l *LibVirt) createDisk(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	newDiskPath := filepath.Join(imagesBasePath, fmt.Sprintf("%s.qcow2", sandboxDto.Id))

	// Get the remote host from libvirt URI
	host := l.extractHostFromURI()
	if host == "" {
		return "", fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	// Check if disk already exists (idempotency - avoid duplicate create attempts)
	checkCmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("test -f %s && echo exists", newDiskPath))
	if output, _ := checkCmd.Output(); strings.TrimSpace(string(output)) == "exists" {
		log.Infof("Disk %s already exists, skipping creation", newDiskPath)
		return newDiskPath, nil
	}

	log.Infof("Creating disk %s from base image %s", newDiskPath, baseImagePath)

	// Create a qcow2 overlay disk with the base image as backing file
	// This uses copy-on-write, so each VM has its own changes without modifying the base
	createDiskCmd := fmt.Sprintf("qemu-img create -f qcow2 -F qcow2 -b %s %s && chown libvirt-qemu:kvm %s",
		baseImagePath, newDiskPath, newDiskPath)

	log.Infof("Executing on remote server: %s", createDiskCmd)

	cmd := exec.CommandContext(ctx, "ssh", host, createDiskCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to create disk: %v, output: %s", err, string(output))
		return "", fmt.Errorf("failed to create disk on remote server: %w (output: %s)", err, string(output))
	}

	log.Infof("Successfully created disk %s", newDiskPath)
	return newDiskPath, nil
}

func (l *LibVirt) copyNVRAM(ctx context.Context, sandboxId string) (string, error) {
	nvramPath := filepath.Join(nvramBasePath, fmt.Sprintf("%s_VARS.fd", sandboxId))

	log.Infof("Copying NVRAM from %s to %s", templateNVRAM, nvramPath)

	host := l.extractHostFromURI()
	if host == "" {
		return "", fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	copyCmd := fmt.Sprintf("cp %s %s && chown libvirt-qemu:kvm %s", templateNVRAM, nvramPath, nvramPath)

	cmd := exec.CommandContext(ctx, "ssh", host, copyCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to copy NVRAM: %v, output: %s", err, string(output))
		return "", fmt.Errorf("failed to copy NVRAM: %w (output: %s)", err, string(output))
	}

	log.Infof("Successfully copied NVRAM to %s", nvramPath)
	return nvramPath, nil
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

// getActualDomainIP gets the actual IP address from DHCP lease (not pre-calculated)
func (l *LibVirt) getActualDomainIP(sandboxId string) string {
	conn, err := l.getConnection()
	if err != nil {
		return ""
	}

	domain, err := conn.LookupDomainByName(sandboxId)
	if err != nil {
		return ""
	}
	defer domain.Free()

	// Get actual IP from DHCP lease
	return l.getDomainIP(conn, domain)
}

// waitForDaemonReady waits for the daemon API to be accessible on the sandbox
// It polls for the actual IP from DHCP lease, not the pre-calculated one
func (l *LibVirt) waitForDaemonReady(ctx context.Context, sandboxId, _ string) error {
	timeout := time.Duration(l.daemonStartTimeoutSec) * time.Second
	deadline := time.Now().Add(timeout)

	log.Infof("Waiting for daemon to be ready on sandbox %s (timeout: %v)", sandboxId, timeout)

	// Create HTTP client with SSH tunnel if remote
	var client *http.Client
	if IsRemoteURI(l.libvirtURI) {
		sshHost := l.extractHostFromURI()
		transport := GetSSHTunnelTransport(sshHost)
		client = &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		}
		log.Infof("Using SSH tunnel via %s to reach daemon", sshHost)
	} else {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastIP string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("daemon not ready after %v", timeout)
			}

			// Get actual IP from domain (not pre-calculated)
			actualIP := l.getActualDomainIP(sandboxId)
			if actualIP == "" {
				log.Debugf("Waiting for domain %s to get IP address...", sandboxId)
				continue
			}
			if actualIP != lastIP {
				log.Infof("Domain %s has IP: %s", sandboxId, actualIP)
				lastIP = actualIP
			}

			daemonURL := fmt.Sprintf("http://%s:2280/version", actualIP)
			resp, err := client.Get(daemonURL)
			if err != nil {
				log.Debugf("Daemon not ready yet at %s: %v", actualIP, err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				log.Infof("Daemon is ready on sandbox %s", sandboxId)
				return nil
			}
			log.Debugf("Daemon returned status %d, waiting...", resp.StatusCode)
		}
	}
}
