// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GpuDevice represents a GPU device available for passthrough
type GpuDevice struct {
	PciAddress  string `json:"pciAddress"`  // e.g., "0000:01:00.0"
	DevicePath  string `json:"devicePath"`  // e.g., "/sys/bus/pci/devices/0000:01:00.0"
	Vendor      string `json:"vendor"`      // e.g., "NVIDIA Corporation"
	Model       string `json:"model"`       // e.g., "GeForce RTX 4090"
	VendorId    string `json:"vendorId"`    // e.g., "10de"
	DeviceId    string `json:"deviceId"`    // e.g., "2684"
	IommuGroup  string `json:"iommuGroup"`  // e.g., "12"
	BoundDriver string `json:"boundDriver"` // e.g., "vfio-pci" or "nvidia"
	IsAvailable bool   `json:"isAvailable"` // True if bound to vfio-pci
}

// ListGpuDevices lists all GPU devices available for passthrough
func (c *Client) ListGpuDevices(ctx context.Context) ([]GpuDevice, error) {
	log.Info("Scanning for GPU devices")

	// Find GPU devices via lspci
	output, err := c.runCommandOutput(ctx, "lspci", "-nn", "-D")
	if err != nil {
		return nil, fmt.Errorf("failed to list PCI devices: %w", err)
	}

	var gpus []GpuDevice

	for _, line := range splitLines(output) {
		// Look for VGA, 3D, and Display controllers (common GPU classes)
		if !strings.Contains(line, "VGA") &&
			!strings.Contains(line, "3D") &&
			!strings.Contains(line, "Display") {
			continue
		}

		// Skip ASPEED (BMC) and other non-passthrough devices
		if strings.Contains(line, "ASPEED") {
			continue
		}

		// Parse PCI address from line start (format: "0000:01:00.0 ...")
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		pciAddr := parts[0]
		devicePath := filepath.Join("/sys/bus/pci/devices", pciAddr)

		gpu := GpuDevice{
			PciAddress: pciAddr,
			DevicePath: devicePath,
		}

		// Extract vendor and device IDs from line (format: "... [xxxx:yyyy]")
		if idx := strings.LastIndex(line, "["); idx != -1 {
			ids := line[idx+1:]
			if endIdx := strings.Index(ids, "]"); endIdx != -1 {
				ids = ids[:endIdx]
				if idParts := strings.Split(ids, ":"); len(idParts) == 2 {
					gpu.VendorId = idParts[0]
					gpu.DeviceId = idParts[1]
				}
			}
		}

		// Get vendor name
		vendorOutput, _ := c.runCommandOutput(ctx, "cat", filepath.Join(devicePath, "vendor"))
		if strings.TrimSpace(vendorOutput) == "0x10de" {
			gpu.Vendor = "NVIDIA Corporation"
		} else if strings.TrimSpace(vendorOutput) == "0x1002" {
			gpu.Vendor = "AMD"
		} else {
			gpu.Vendor = strings.TrimSpace(vendorOutput)
		}

		// Parse model from lspci output
		if colonIdx := strings.Index(parts[1], ":"); colonIdx != -1 {
			gpu.Model = strings.TrimSpace(parts[1][colonIdx+1:])
			// Remove the [xxxx:yyyy] suffix
			if bracketIdx := strings.LastIndex(gpu.Model, "["); bracketIdx != -1 {
				gpu.Model = strings.TrimSpace(gpu.Model[:bracketIdx])
			}
		}

		// Get IOMMU group
		iommuLink, _ := c.runCommandOutput(ctx, "readlink", "-f", filepath.Join(devicePath, "iommu_group"))
		gpu.IommuGroup = filepath.Base(strings.TrimSpace(iommuLink))

		// Get current driver
		driverLink, _ := c.runCommandOutput(ctx, "readlink", "-f", filepath.Join(devicePath, "driver"))
		gpu.BoundDriver = filepath.Base(strings.TrimSpace(driverLink))

		// Check if available (bound to vfio-pci)
		gpu.IsAvailable = gpu.BoundDriver == "vfio-pci"

		gpus = append(gpus, gpu)
	}

	log.Infof("Found %d GPU devices", len(gpus))
	return gpus, nil
}

// BindGpuToVfio binds a GPU to the vfio-pci driver for passthrough
func (c *Client) BindGpuToVfio(ctx context.Context, pciAddress string) error {
	log.Infof("Binding GPU %s to vfio-pci", pciAddress)

	devicePath := filepath.Join("/sys/bus/pci/devices", pciAddress)

	// Get vendor and device IDs
	vendorId, err := c.runCommandOutput(ctx, "cat", filepath.Join(devicePath, "vendor"))
	if err != nil {
		return fmt.Errorf("failed to read vendor ID: %w", err)
	}
	deviceId, err := c.runCommandOutput(ctx, "cat", filepath.Join(devicePath, "device"))
	if err != nil {
		return fmt.Errorf("failed to read device ID: %w", err)
	}

	vendorId = strings.TrimSpace(strings.TrimPrefix(vendorId, "0x"))
	deviceId = strings.TrimSpace(strings.TrimPrefix(deviceId, "0x"))

	// Unbind from current driver
	driverLink, _ := c.runCommandOutput(ctx, "readlink", "-f", filepath.Join(devicePath, "driver"))
	currentDriver := filepath.Base(strings.TrimSpace(driverLink))

	if currentDriver != "" && currentDriver != "vfio-pci" {
		log.Infof("Unbinding from current driver: %s", currentDriver)
		unbindPath := filepath.Join(devicePath, "driver", "unbind")
		if err := c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", pciAddress, unbindPath)); err != nil {
			return fmt.Errorf("failed to unbind from driver: %w", err)
		}
	}

	// Add device ID to vfio-pci
	newIdPath := "/sys/bus/pci/drivers/vfio-pci/new_id"
	if err := c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s %s' > %s", vendorId, deviceId, newIdPath)); err != nil {
		log.Warnf("Failed to add new_id (may already exist): %v", err)
	}

	// Bind to vfio-pci
	bindPath := "/sys/bus/pci/drivers/vfio-pci/bind"
	if err := c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", pciAddress, bindPath)); err != nil {
		return fmt.Errorf("failed to bind to vfio-pci: %w", err)
	}

	log.Infof("GPU %s bound to vfio-pci successfully", pciAddress)
	return nil
}

// UnbindGpuFromVfio unbinds a GPU from vfio-pci
func (c *Client) UnbindGpuFromVfio(ctx context.Context, pciAddress string) error {
	log.Infof("Unbinding GPU %s from vfio-pci", pciAddress)

	unbindPath := "/sys/bus/pci/drivers/vfio-pci/unbind"
	if err := c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", pciAddress, unbindPath)); err != nil {
		return fmt.Errorf("failed to unbind from vfio-pci: %w", err)
	}

	// Trigger driver probe to rebind to original driver
	probePath := filepath.Join("/sys/bus/pci/devices", pciAddress, "driver_override")
	if err := c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '' > %s", probePath)); err != nil {
		log.Warnf("Failed to clear driver_override: %v", err)
	}

	driverProbePath := "/sys/bus/pci/drivers_probe"
	if err := c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", pciAddress, driverProbePath)); err != nil {
		log.Warnf("Failed to trigger driver probe: %v", err)
	}

	log.Infof("GPU %s unbound from vfio-pci", pciAddress)
	return nil
}

// AddGpuToVm hot-adds a GPU to a running VM
func (c *Client) AddGpuToVm(ctx context.Context, sandboxId, pciAddress string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Adding GPU %s to sandbox %s", pciAddress, sandboxId)

	devicePath := filepath.Join("/sys/bus/pci/devices", pciAddress)

	// Ensure device is bound to vfio-pci
	driverLink, _ := c.runCommandOutput(ctx, "readlink", "-f", filepath.Join(devicePath, "driver"))
	currentDriver := filepath.Base(strings.TrimSpace(driverLink))

	if currentDriver != "vfio-pci" {
		return fmt.Errorf("GPU %s is not bound to vfio-pci (current: %s)", pciAddress, currentDriver)
	}

	// Add device via CH API
	deviceConfig := DeviceConfig{
		Path:  devicePath,
		Iommu: true,
		Id:    fmt.Sprintf("gpu-%s", strings.ReplaceAll(pciAddress, ":", "-")),
	}

	deviceConfigStr := fmt.Sprintf("path=%s,iommu=on,id=%s", devicePath, deviceConfig.Id)

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.add-device", deviceConfigStr); err != nil {
		return fmt.Errorf("failed to add GPU to VM: %w", err)
	}

	log.Infof("GPU %s added to sandbox %s", pciAddress, sandboxId)
	return nil
}

// RemoveGpuFromVm hot-removes a GPU from a running VM
func (c *Client) RemoveGpuFromVm(ctx context.Context, sandboxId, deviceId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Removing GPU %s from sandbox %s", deviceId, sandboxId)

	removeConfig := map[string]string{
		"id": deviceId,
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.remove-device", removeConfig); err != nil {
		return fmt.Errorf("failed to remove GPU from VM: %w", err)
	}

	log.Infof("GPU %s removed from sandbox %s", deviceId, sandboxId)
	return nil
}

// GetVmGpuDevices returns the GPU devices attached to a VM
func (c *Client) GetVmGpuDevices(ctx context.Context, sandboxId string) ([]DeviceConfig, error) {
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return nil, err
	}

	if info.Config == nil {
		return nil, nil
	}

	return info.Config.Devices, nil
}
