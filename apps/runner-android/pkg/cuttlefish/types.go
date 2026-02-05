// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"time"
)

// InstanceState represents the current state of a Cuttlefish instance
type InstanceState string

const (
	InstanceStateRunning  InstanceState = "Running"
	InstanceStateStopped  InstanceState = "Stopped"
	InstanceStateStarting InstanceState = "Starting"
	InstanceStateStopping InstanceState = "Stopping"
	InstanceStateUnknown  InstanceState = "Unknown"
)

// ClientConfig holds configuration for the Cuttlefish client
type ClientConfig struct {
	// InstancesPath is the base directory for instance data
	InstancesPath string
	// ArtifactsPath is the directory containing Cuttlefish artifacts (system images, etc.)
	ArtifactsPath string
	// CVDHome is the home directory for Cuttlefish (contains runtime files)
	CVDHome string
	// LaunchCVDPath is the path to the launch_cvd binary
	LaunchCVDPath string
	// StopCVDPath is the path to the stop_cvd binary
	StopCVDPath string
	// CVDPath is the path to the cvd binary
	CVDPath string
	// ADBPath is the path to the adb binary
	ADBPath string
	// DefaultCpus is the default number of vCPUs
	DefaultCpus int
	// DefaultMemoryMB is the default memory in MB
	DefaultMemoryMB uint64
	// DefaultDiskGB is the default disk size in GB
	DefaultDiskGB int
	// BaseInstanceNum is the starting instance number for allocation
	BaseInstanceNum int
	// MaxInstances is the maximum number of concurrent instances
	MaxInstances int
	// SSHHost is the remote host for SSH-based operations (empty for local mode)
	SSHHost string
	// SSHKeyPath is the path to the SSH private key
	SSHKeyPath string
	// ADBBasePort is the base port for ADB connections (instance port = base + instance_num - 1)
	ADBBasePort int
	// WebRTCBasePort is the base port for WebRTC streaming
	WebRTCBasePort int
}

// InstanceInfo represents information about a Cuttlefish instance
type InstanceInfo struct {
	// SandboxId is the unique identifier for this sandbox
	SandboxId string `json:"sandboxId"`
	// InstanceNum is the Cuttlefish instance number (1-based)
	InstanceNum int `json:"instanceNum"`
	// State is the current instance state
	State InstanceState `json:"state"`
	// Cpus is the number of vCPUs allocated
	Cpus int `json:"cpus"`
	// MemoryMB is the memory allocated in MB
	MemoryMB uint64 `json:"memoryMB"`
	// DiskGB is the disk size in GB
	DiskGB int `json:"diskGB"`
	// ADBPort is the ADB port for this instance
	ADBPort int `json:"adbPort"`
	// ADBSerial is the ADB serial string (e.g., "0.0.0.0:6520")
	ADBSerial string `json:"adbSerial"`
	// WebRTCPort is the WebRTC streaming port
	WebRTCPort int `json:"webrtcPort"`
	// CreatedAt is when the instance was created
	CreatedAt time.Time `json:"createdAt"`
	// RuntimeDir is the instance runtime directory
	RuntimeDir string `json:"runtimeDir"`
	// Metadata is custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SandboxInfo represents a sandbox instance managed by runner-android
// This maintains compatibility with the executor interface
type SandboxInfo struct {
	Id        string            `json:"id"`
	State     InstanceState     `json:"state"`
	Vcpus     int               `json:"vcpus"`
	MemoryMB  uint64            `json:"memoryMB"`
	ADBSerial string            `json:"adbSerial,omitempty"`
	ADBPort   int               `json:"adbPort,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateOptions specifies options for creating a new Cuttlefish instance
type CreateOptions struct {
	SandboxId string
	Cpus      int
	MemoryMB  uint64
	DiskGB    int
	Snapshot  string            // Android system image variant (optional)
	Metadata  map[string]string // Custom metadata
}

// InstanceMapping stores the mapping between sandbox IDs and instance numbers
type InstanceMapping struct {
	SandboxId   string    `json:"sandboxId"`
	InstanceNum int       `json:"instanceNum"`
	CreatedAt   time.Time `json:"createdAt"`
}

// RemoteMetrics represents system metrics from a remote host
type RemoteMetrics struct {
	CPUUsagePercent    float64 `json:"cpuUsagePercent"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent"`
	DiskUsagePercent   float64 `json:"diskUsagePercent"`
	TotalCPUs          int     `json:"totalCPUs"`
	TotalMemoryGiB     float64 `json:"totalMemoryGiB"`
	TotalDiskGiB       float64 `json:"totalDiskGiB"`
}

// CVDStatus represents the status output from cvd status command
type CVDStatus struct {
	Instances []CVDInstanceStatus `json:"instances"`
}

// CVDInstanceStatus represents status of a single CVD instance
type CVDInstanceStatus struct {
	InstanceNum int    `json:"instance_num"`
	State       string `json:"state"`
	ADBSerial   string `json:"adb_serial"`
	WebRTCPort  int    `json:"webrtc_port"`
}
