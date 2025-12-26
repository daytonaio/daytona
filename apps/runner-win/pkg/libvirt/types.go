// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"time"

	"libvirt.org/go/libvirt"
)

// DomainState represents the state of a libvirt domain/VM
type DomainState libvirt.DomainState

const (
	DomainStateNoState     DomainState = DomainState(libvirt.DOMAIN_NOSTATE)
	DomainStateRunning     DomainState = DomainState(libvirt.DOMAIN_RUNNING)
	DomainStateBlocked     DomainState = DomainState(libvirt.DOMAIN_BLOCKED)
	DomainStatePaused      DomainState = DomainState(libvirt.DOMAIN_PAUSED)
	DomainStateShutdown    DomainState = DomainState(libvirt.DOMAIN_SHUTDOWN)
	DomainStateShutoff     DomainState = DomainState(libvirt.DOMAIN_SHUTOFF)
	DomainStateCrashed     DomainState = DomainState(libvirt.DOMAIN_CRASHED)
	DomainStatePMSuspended DomainState = DomainState(libvirt.DOMAIN_PMSUSPENDED)
)

// DomainSummary represents a summary of a libvirt domain/VM
type DomainSummary struct {
	UUID   string
	Name   string
	State  DomainState
	Memory uint64 // in KiB
	VCPUs  uint
	ID     int // Domain ID (-1 if not running)
}

// DomainInfo represents detailed information about a domain/VM
type DomainInfo struct {
	UUID         string
	Name         string
	State        DomainState
	Memory       uint64 // in KiB
	MaxMemory    uint64 // in KiB
	VCPUs        uint
	CPUTime      uint64
	ID           int
	Created      time.Time
	OSType       string
	Architecture string
	Metadata     map[string]string
}

// SystemInfo represents system-wide information from libvirt
type SystemInfo struct {
	Hostname        string
	HypervisorType  string
	LibvirtVersion  uint64
	ConnectionURI   string
	TotalMemory     uint64 // in KiB
	TotalCPUs       int
	DomainsActive   int
	DomainsInactive int
	DomainsTotal    int
}

// DomainListOptions represents options for listing domains
type DomainListOptions struct {
	All      bool
	Active   bool
	Inactive bool
}

// Resources represents VM resource limits
type Resources struct {
	Memory    uint64 // in KiB
	MaxMemory uint64 // in KiB
	VCPUs     uint
	CPUShares uint64
}

// NetworkInterface represents a network interface configuration
type NetworkInterface struct {
	Type       string // network, bridge, direct
	Source     string // network/bridge name
	Model      string // virtio, e1000, etc.
	MACAddress string
	IPAddress  string
}

// DiskDevice represents a disk device configuration
type DiskDevice struct {
	Type   string // file, block, volume
	Device string // disk, cdrom, floppy
	Source string // path to disk image
	Target string // vda, hda, etc.
	Bus    string // virtio, ide, scsi, etc.
	Format string // qcow2, raw, etc.
}

// DomainConfig represents domain configuration
type DomainConfig struct {
	Name         string
	UUID         string
	Memory       uint64 // in KiB
	MaxMemory    uint64 // in KiB
	VCPUs        uint
	OSType       string // hvm, linux, etc.
	Architecture string // x86_64, aarch64, etc.
	Machine      string // pc-q35-6.2, etc.
	Emulator     string // path to qemu binary
	Disks        []DiskDevice
	Networks     []NetworkInterface
	Metadata     map[string]string
}
