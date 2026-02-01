// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/daytonaio/runner-win/pkg/models/enums"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirtxml"
)

// CloneVMInfo holds information about a cloned VM
type CloneVMInfo struct {
	Id            string
	State         string
	DaemonVersion string
}

// CloneVM creates a completely independent copy of a VM with flattened filesystem
// Unlike fork/linked clone which uses qcow2 backing file, CloneVM creates a standalone disk
// This means:
// - Disk is a complete copy (no backing file dependency)
// - Takes longer than linked clone due to disk copy
// - Cloned VM is completely independent of source
//
// The clone process:
// 1. Pause source VM (if running)
// 2. Flatten disk: qemu-img convert to create standalone qcow2
// 3. Resume source VM
// 4. Copy NVRAM
// 5. Create new VM with flattened disk (fresh boot)
func (l *LibVirt) CloneVM(ctx context.Context, sourceSandboxId, newSandboxId string, sourceStopped bool) (*CloneVMInfo, error) {
	sourceMutex := l.getDomainMutex(sourceSandboxId)
	sourceMutex.Lock()
	defer sourceMutex.Unlock()

	newMutex := l.getDomainMutex(newSandboxId)
	newMutex.Lock()
	defer newMutex.Unlock()

	log.Infof("Cloning sandbox %s to %s (complete disk copy, sourceStopped: %v)", sourceSandboxId, newSandboxId, sourceStopped)

	conn, err := l.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Check if target domain already exists
	if existingDomain, err := conn.LookupDomainByName(newSandboxId); err == nil && existingDomain != nil {
		existingDomain.Free()
		return nil, fmt.Errorf("target sandbox %s already exists", newSandboxId)
	}

	// Verify the source domain exists
	sourceDomain, err := conn.LookupDomainByName(sourceSandboxId)
	if err != nil {
		return nil, fmt.Errorf("source sandbox not found: %w", err)
	}
	defer sourceDomain.Free()

	// Get source domain state
	state, _, err := sourceDomain.GetState()
	if err != nil {
		return nil, fmt.Errorf("failed to get source domain state: %w", err)
	}

	// Pause source VM if running for consistent disk copy
	wasRunning := state == 1 // VIR_DOMAIN_RUNNING
	if wasRunning && !sourceStopped {
		// Before pausing, configure Windows to prevent "unexpected shutdown" dialog on cloned VM
		// We need to set multiple settings to fully suppress the dialog
		log.Infof("Configuring source VM %s to prevent unexpected shutdown dialog on clone", sourceSandboxId)

		// 1. Set boot policy to ignore all failures
		bootPolicyCmd := "bcdedit /set {current} bootstatuspolicy ignoreallfailures"
		if resp, err := l.ExecuteGuestCommand(ctx, sourceSandboxId, bootPolicyCmd, 30); err != nil {
			log.Warnf("Failed to set boot policy (continuing anyway): %v", err)
		} else if resp.ExitCode != 0 {
			log.Warnf("Boot policy command returned exit code %d: %s (continuing anyway)", resp.ExitCode, resp.Result)
		} else {
			log.Infof("Successfully set boot policy to ignore failures")
		}

		// 2. Disable automatic recovery
		recoveryCmd := "bcdedit /set {current} recoveryenabled No"
		if resp, err := l.ExecuteGuestCommand(ctx, sourceSandboxId, recoveryCmd, 30); err != nil {
			log.Warnf("Failed to disable recovery (continuing anyway): %v", err)
		} else if resp.ExitCode != 0 {
			log.Warnf("Recovery disable command returned exit code %d: %s (continuing anyway)", resp.ExitCode, resp.Result)
		} else {
			log.Infof("Successfully disabled recovery")
		}

		// 3. Disable Shutdown Event Tracker via Group Policy registry
		// This is the most reliable way to prevent the "unexpected shutdown" dialog
		shutdownTrackerCmd := `powershell -Command "New-Item -Path 'HKLM:\SOFTWARE\Policies\Microsoft\Windows NT\Reliability' -Force | Out-Null; New-ItemProperty -Path 'HKLM:\SOFTWARE\Policies\Microsoft\Windows NT\Reliability' -Name 'ShutdownReasonOn' -Value 0 -PropertyType DWord -Force | Out-Null"`
		if resp, err := l.ExecuteGuestCommand(ctx, sourceSandboxId, shutdownTrackerCmd, 30); err != nil {
			log.Warnf("Failed to disable shutdown tracker (continuing anyway): %v", err)
		} else if resp.ExitCode != 0 {
			log.Warnf("Shutdown tracker command returned exit code %d: %s (continuing anyway)", resp.ExitCode, resp.Result)
		} else {
			log.Infof("Successfully disabled Shutdown Event Tracker")
		}

		// 4. Also disable it for servers (ShutdownReasonUI)
		shutdownUICmd := `powershell -Command "New-ItemProperty -Path 'HKLM:\SOFTWARE\Policies\Microsoft\Windows NT\Reliability' -Name 'ShutdownReasonUI' -Value 0 -PropertyType DWord -Force | Out-Null"`
		if resp, err := l.ExecuteGuestCommand(ctx, sourceSandboxId, shutdownUICmd, 30); err != nil {
			log.Warnf("Failed to disable shutdown UI (continuing anyway): %v", err)
		} else if resp.ExitCode != 0 {
			log.Warnf("Shutdown UI command returned exit code %d: %s (continuing anyway)", resp.ExitCode, resp.Result)
		} else {
			log.Infof("Successfully disabled Shutdown Reason UI")
		}

		// 5. Clear the dirty shutdown registry key to prevent "Windows was not shut down properly" dialog
		clearDirtyCmd := `reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Reliability" /v DirtyShutdown /t REG_DWORD /d 0 /f`
		if resp, err := l.ExecuteGuestCommand(ctx, sourceSandboxId, clearDirtyCmd, 30); err != nil {
			log.Warnf("Failed to clear dirty shutdown flag (continuing anyway): %v", err)
		} else if resp.ExitCode != 0 {
			log.Warnf("Clear dirty shutdown command returned exit code %d: %s (continuing anyway)", resp.ExitCode, resp.Result)
		} else {
			log.Infof("Successfully cleared dirty shutdown flag")
		}

		// 6. Flush filesystem to ensure changes are written to disk before we pause
		flushCmd := "powershell -Command \"Write-VolumeCache -DriveLetter C\""
		if resp, err := l.ExecuteGuestCommand(ctx, sourceSandboxId, flushCmd, 30); err != nil {
			log.Warnf("Failed to flush filesystem (continuing anyway): %v", err)
		} else {
			log.Infof("Filesystem flush completed (exit code: %d)", resp.ExitCode)
		}

		log.Infof("Pausing source VM %s for clone disk copy", sourceSandboxId)
		if err := sourceDomain.Suspend(); err != nil {
			return nil, fmt.Errorf("failed to pause source VM: %w", err)
		}
		// Ensure we resume on any error or success
		defer func() {
			log.Infof("Resuming source VM %s after clone disk copy", sourceSandboxId)
			if err := sourceDomain.Resume(); err != nil {
				log.Warnf("Failed to resume source VM: %v", err)
			}
		}()
	}

	// Get source domain XML to use as template and extract disk path
	sourceXML, err := sourceDomain.GetXMLDesc(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get source domain XML: %w", err)
	}

	// Parse the XML to get the source disk path
	var domainCfg libvirtxml.Domain
	if err := domainCfg.Unmarshal(sourceXML); err != nil {
		return nil, fmt.Errorf("failed to parse source domain XML: %w", err)
	}

	// Extract source disk path from XML
	var sourceDiskPath string
	if domainCfg.Devices != nil && len(domainCfg.Devices.Disks) > 0 {
		for _, disk := range domainCfg.Devices.Disks {
			if disk.Device == "disk" && disk.Source != nil && disk.Source.File != nil {
				sourceDiskPath = disk.Source.File.File
				break
			}
		}
	}
	if sourceDiskPath == "" {
		return nil, fmt.Errorf("could not find source disk path in domain XML")
	}

	// Derive target paths from source disk path
	// Source: /path/to/vms/sourceSandboxId/sourceSandboxId.qcow2
	// Target: /path/to/vms/newSandboxId/newSandboxId.qcow2
	sourceVMDir := filepath.Dir(sourceDiskPath)
	baseVMPath := filepath.Dir(sourceVMDir)
	targetVMPath := filepath.Join(baseVMPath, newSandboxId)
	targetDiskPath := filepath.Join(targetVMPath, fmt.Sprintf("%s.qcow2", newSandboxId))

	// Determine if running locally or remotely
	isLocal := l.isLocalURI()
	host := ""
	if !isLocal {
		host = l.extractHostFromURI()
		if host == "" {
			return nil, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
		}
	}

	// Create target VM directory
	var mkdirCmd *exec.Cmd
	if isLocal {
		mkdirCmd = exec.CommandContext(ctx, "mkdir", "-p", targetVMPath)
	} else {
		mkdirCmd = exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("mkdir -p %s", targetVMPath))
	}
	if err := mkdirCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create target VM directory: %w", err)
	}

	// Copy the overlay disk file directly (preserves backing file reference)
	// This is MUCH faster than flattening because we only copy the delta layer (~3GB)
	// instead of the full flattened disk (~15GB+)
	// The clone will share the same read-only base image as the source
	log.Infof("Copying overlay disk for clone (fast copy, preserves backing file)")
	copyCmdStr := fmt.Sprintf("cp %s %s", sourceDiskPath, targetDiskPath)
	var copyCmd *exec.Cmd
	if isLocal {
		log.Infof("Executing locally: %s", copyCmdStr)
		copyCmd = exec.CommandContext(ctx, "bash", "-c", copyCmdStr)
	} else {
		log.Infof("Executing on remote server %s: %s", host, copyCmdStr)
		copyCmd = exec.CommandContext(ctx, "ssh", host, copyCmdStr)
	}
	if output, err := copyCmd.CombinedOutput(); err != nil {
		l.cleanupDomainFiles(ctx, newSandboxId)
		return nil, fmt.Errorf("failed to copy disk: %w (output: %s)", err, string(output))
	}

	log.Infof("Disk copy completed for clone")

	// Copy NVRAM for UEFI boot
	nvramPath, err := l.copyNVRAM(ctx, newSandboxId)
	if err != nil {
		l.cleanupDomainFiles(ctx, newSandboxId)
		return nil, fmt.Errorf("failed to copy NVRAM: %w", err)
	}

	// Generate MAC and IP for the new VM
	mac := GenerateMACFromSandboxID(newSandboxId)
	ip := CalculateIPFromSandboxID(newSandboxId)
	log.Infof("Generated MAC=%s IP=%s for cloned sandbox %s", mac, ip, newSandboxId)

	// Add DHCP reservation
	if err := l.AddDHCPReservation(mac, ip, newSandboxId); err != nil {
		log.Warnf("Failed to add DHCP reservation: %v (continuing anyway)", err)
	}

	// Modify the XML for the new VM (update name, disk path, MAC, NVRAM)
	newXML, err := l.modifyXMLForClone(sourceXML, newSandboxId, targetDiskPath, nvramPath, mac)
	if err != nil {
		l.cleanupDomainFiles(ctx, newSandboxId)
		return nil, fmt.Errorf("failed to modify domain XML: %w", err)
	}

	// Define the new domain
	log.Infof("Defining cloned domain %s", newSandboxId)
	newDomain, err := conn.DomainDefineXML(newXML)
	if err != nil {
		l.cleanupDomainFiles(ctx, newSandboxId)
		return nil, fmt.Errorf("failed to define cloned domain: %w", err)
	}
	defer newDomain.Free()

	// Start the new domain
	log.Infof("Starting cloned domain %s", newSandboxId)
	if err := newDomain.Create(); err != nil {
		// Clean up the defined domain on start failure
		newDomain.Undefine()
		l.cleanupDomainFiles(ctx, newSandboxId)
		return nil, fmt.Errorf("failed to start cloned domain: %w", err)
	}

	// Update states cache
	if l.statesCache != nil {
		l.statesCache.SetSandboxState(ctx, newSandboxId, enums.SandboxStateStarted)
	}

	// Wait for the VM to get its IP and be ready
	if err := l.waitForVMReady(ctx, newSandboxId, 60*time.Second); err != nil {
		log.Warnf("VM ready check failed for %s: %v (continuing anyway)", newSandboxId, err)
	}

	// Get daemon version from the cloned VM
	daemonVersion, _ := l.GetDaemonVersion(ctx, newSandboxId)

	log.Infof("Clone completed: %s -> %s (independent disk)", sourceSandboxId, newSandboxId)

	return &CloneVMInfo{
		Id:            newSandboxId,
		State:         string(enums.SandboxStateStarted),
		DaemonVersion: daemonVersion,
	}, nil
}

// modifyXMLForClone updates the domain XML for the cloned VM
func (l *LibVirt) modifyXMLForClone(sourceXML, newName, diskPath, nvramPath, mac string) (string, error) {
	// Parse the source XML
	var domainCfg libvirtxml.Domain
	if err := domainCfg.Unmarshal(sourceXML); err != nil {
		return "", fmt.Errorf("failed to parse source XML: %w", err)
	}

	// 1. Update domain name
	domainCfg.Name = newName

	// 2. Remove UUID (let libvirt generate new one)
	domainCfg.UUID = ""

	// 3. Update disk path (first disk is the main OS disk)
	if domainCfg.Devices != nil && len(domainCfg.Devices.Disks) > 0 {
		for i := range domainCfg.Devices.Disks {
			if domainCfg.Devices.Disks[i].Device == "disk" && domainCfg.Devices.Disks[i].Source != nil {
				if domainCfg.Devices.Disks[i].Source.File != nil {
					domainCfg.Devices.Disks[i].Source.File.File = diskPath
					break
				}
			}
		}
	}

	// 4. Update MAC address
	if domainCfg.Devices != nil && len(domainCfg.Devices.Interfaces) > 0 {
		if domainCfg.Devices.Interfaces[0].MAC != nil {
			domainCfg.Devices.Interfaces[0].MAC.Address = mac
		}
	}

	// 5. Update NVRAM path
	if domainCfg.OS != nil && domainCfg.OS.NVRam != nil {
		domainCfg.OS.NVRam.NVRam = nvramPath
	}

	// Marshal back to XML
	newXML, err := domainCfg.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified XML: %w", err)
	}

	return newXML, nil
}

// waitForVMReady waits for the VM to be ready (get IP and respond to health check)
func (l *LibVirt) waitForVMReady(ctx context.Context, sandboxId string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Try to get daemon version as health check
		_, err := l.GetDaemonVersion(ctx, sandboxId)
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("VM not ready after %v", timeout)
}
