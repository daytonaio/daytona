// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"time"
)

// VmState represents the current state of a Cloud Hypervisor VM
type VmState string

const (
	VmStateCreated    VmState = "Created"
	VmStateRunning    VmState = "Running"
	VmStatePaused     VmState = "Paused"
	VmStateShutdown   VmState = "Shutdown"
	VmStateNotCreated VmState = "NotCreated"
)

// VmConfig represents the full VM configuration for Cloud Hypervisor
type VmConfig struct {
	Payload         *PayloadConfig   `json:"payload,omitempty"`
	Disks           []DiskConfig     `json:"disks,omitempty"`
	Net             []NetConfig      `json:"net,omitempty"`
	Cpus            *CpusConfig      `json:"cpus,omitempty"`
	Memory          *MemoryConfig    `json:"memory,omitempty"`
	Serial          *ConsoleConfig   `json:"serial,omitempty"`
	Console         *ConsoleConfig   `json:"console,omitempty"`
	Devices         []DeviceConfig   `json:"devices,omitempty"`
	Vsock           *VsockConfig     `json:"vsock,omitempty"`
	Rng             *RngConfig       `json:"rng,omitempty"`
	Balloon         *BalloonConfig   `json:"balloon,omitempty"`
	Fs              []FsConfig       `json:"fs,omitempty"`
	Pmem            []PmemConfig     `json:"pmem,omitempty"`
	Iommu           bool             `json:"iommu,omitempty"`
	Watchdog        bool             `json:"watchdog,omitempty"`
	Platform        *PlatformConfig  `json:"platform,omitempty"`
	Numa            []NumaConfig     `json:"numa,omitempty"`
	RateLimitGroups []RateLimitGroup `json:"rate_limit_groups,omitempty"`
}

// PayloadConfig specifies the kernel/firmware to boot
type PayloadConfig struct {
	Firmware  string `json:"firmware,omitempty"`  // Path to hypervisor-fw or OVMF
	Kernel    string `json:"kernel,omitempty"`    // Path to vmlinux for direct kernel boot
	Cmdline   string `json:"cmdline,omitempty"`   // Kernel command line
	Initramfs string `json:"initramfs,omitempty"` // Path to initramfs
}

// DiskConfig represents a block device configuration
type DiskConfig struct {
	Path           string          `json:"path"`
	Readonly       bool            `json:"readonly,omitempty"`
	Direct         bool            `json:"direct,omitempty"`
	Iommu          bool            `json:"iommu,omitempty"`
	NumQueues      int             `json:"num_queues,omitempty"`
	QueueSize      int             `json:"queue_size,omitempty"`
	VhostUser      bool            `json:"vhost_user,omitempty"`
	VhostSocket    string          `json:"vhost_socket,omitempty"`
	RateLimitGroup string          `json:"rate_limit_group,omitempty"`
	RateLimiter    *RateLimiter    `json:"rate_limiter_config,omitempty"`
	Id             string          `json:"id,omitempty"`
	DisableIoUring bool            `json:"disable_io_uring,omitempty"`
	DisableAio     bool            `json:"disable_aio,omitempty"`
	PciSegment     int             `json:"pci_segment,omitempty"`
	Serial         string          `json:"serial,omitempty"`
	QueueAffinity  []QueueAffinity `json:"queue_affinity,omitempty"`
}

// NetConfig represents a network device configuration
type NetConfig struct {
	Tap         string       `json:"tap,omitempty"`
	Ip          string       `json:"ip,omitempty"`
	Mask        string       `json:"mask,omitempty"`
	Mac         string       `json:"mac,omitempty"`
	Fds         []int        `json:"fds,omitempty"`
	Iommu       bool         `json:"iommu,omitempty"`
	NumQueues   int          `json:"num_queues,omitempty"`
	QueueSize   int          `json:"queue_size,omitempty"`
	Id          string       `json:"id,omitempty"`
	VhostUser   bool         `json:"vhost_user,omitempty"`
	VhostSocket string       `json:"socket,omitempty"`
	VhostMode   string       `json:"vhost_mode,omitempty"` // "client" or "server"
	RateLimiter *RateLimiter `json:"rate_limiter_config,omitempty"`
	PciSegment  int          `json:"pci_segment,omitempty"`
	OffloadTso  bool         `json:"offload_tso,omitempty"`
	OffloadUfo  bool         `json:"offload_ufo,omitempty"`
	OffloadCsum bool         `json:"offload_csum,omitempty"`
}

// CpusConfig represents CPU configuration
type CpusConfig struct {
	BootVcpus   int           `json:"boot_vcpus"`
	MaxVcpus    int           `json:"max_vcpus"`
	Topology    *CpuTopology  `json:"topology,omitempty"`
	KvmHyperv   bool          `json:"kvm_hyperv,omitempty"`
	MaxPhysBits int           `json:"max_phys_bits,omitempty"`
	Affinity    []CpuAffinity `json:"affinity,omitempty"`
	Features    *CpuFeatures  `json:"features,omitempty"`
	Nested      bool          `json:"nested,omitempty"`
}

// CpuTopology represents CPU topology configuration
type CpuTopology struct {
	ThreadsPerCore int `json:"threads_per_core,omitempty"`
	CoresPerDie    int `json:"cores_per_die,omitempty"`
	DiesPerPackage int `json:"dies_per_package,omitempty"`
	Packages       int `json:"packages,omitempty"`
}

// CpuAffinity represents CPU affinity configuration
type CpuAffinity struct {
	Vcpu     int   `json:"vcpu"`
	HostCpus []int `json:"host_cpus"`
}

// CpuFeatures represents CPU feature flags
type CpuFeatures struct {
	Amx bool `json:"amx,omitempty"`
}

// MemoryConfig represents memory configuration
type MemoryConfig struct {
	Size           uint64       `json:"size"`                      // Memory size in bytes
	Mergeable      bool         `json:"mergeable,omitempty"`       // Enable KSM
	HotplugMethod  string       `json:"hotplug_method,omitempty"`  // "Acpi" or "VirtioMem"
	HotplugSize    *uint64      `json:"hotplug_size,omitempty"`    // Max hotplug size
	HotpluggedSize *uint64      `json:"hotplugged_size,omitempty"` // Current hotplugged size
	Shared         bool         `json:"shared,omitempty"`          // Shared memory for vhost-user
	Hugepages      bool         `json:"hugepages,omitempty"`       // Use huge pages
	HugepageSize   *uint64      `json:"hugepage_size,omitempty"`   // Huge page size
	Prefault       bool         `json:"prefault,omitempty"`        // Prefault pages
	Zones          []MemoryZone `json:"zones,omitempty"`           // NUMA memory zones
	Thp            bool         `json:"thp,omitempty"`             // Transparent huge pages
}

// MemoryZone represents a NUMA memory zone
type MemoryZone struct {
	Id             string  `json:"id"`
	Size           uint64  `json:"size"`
	File           string  `json:"file,omitempty"`
	Mergeable      bool    `json:"mergeable,omitempty"`
	Shared         bool    `json:"shared,omitempty"`
	Hugepages      bool    `json:"hugepages,omitempty"`
	HugepageSize   *uint64 `json:"hugepage_size,omitempty"`
	HostNumaNode   *int    `json:"host_numa_node,omitempty"`
	HotplugSize    *uint64 `json:"hotplug_size,omitempty"`
	HotpluggedSize *uint64 `json:"hotplugged_size,omitempty"`
	Prefault       bool    `json:"prefault,omitempty"`
}

// ConsoleConfig represents serial/console configuration
type ConsoleConfig struct {
	File   string `json:"file,omitempty"`
	Mode   string `json:"mode,omitempty"` // "Off", "Pty", "Tty", "File", "Socket"
	Iommu  bool   `json:"iommu,omitempty"`
	Socket string `json:"socket,omitempty"`
}

// DeviceConfig represents a VFIO device (for GPU passthrough)
type DeviceConfig struct {
	Path       string `json:"path"` // PCI device path (e.g., /sys/bus/pci/devices/0000:01:00.0)
	Iommu      bool   `json:"iommu,omitempty"`
	Id         string `json:"id,omitempty"`
	PciSegment int    `json:"pci_segment,omitempty"`
}

// VsockConfig represents virtio-vsock configuration
type VsockConfig struct {
	Cid    uint64 `json:"cid"`
	Socket string `json:"socket"`
	Iommu  bool   `json:"iommu,omitempty"`
	Id     string `json:"id,omitempty"`
}

// RngConfig represents the random number generator configuration
type RngConfig struct {
	Src   string `json:"src,omitempty"` // Default: /dev/urandom
	Iommu bool   `json:"iommu,omitempty"`
}

// BalloonConfig represents memory balloon configuration
type BalloonConfig struct {
	Size              uint64 `json:"size"`
	DeflateOnOom      bool   `json:"deflate_on_oom,omitempty"`
	FreePageReporting bool   `json:"free_page_reporting,omitempty"`
}

// FsConfig represents virtio-fs configuration
type FsConfig struct {
	Tag       string `json:"tag"`
	Socket    string `json:"socket"`
	NumQueues int    `json:"num_queues,omitempty"`
	QueueSize int    `json:"queue_size,omitempty"`
	Id        string `json:"id,omitempty"`
}

// PmemConfig represents persistent memory configuration
type PmemConfig struct {
	File       string `json:"file"`
	Size       uint64 `json:"size,omitempty"`
	Iommu      bool   `json:"iommu,omitempty"`
	MergeFules bool   `json:"merge_rules,omitempty"`
	Id         string `json:"id,omitempty"`
	PciSegment int    `json:"pci_segment,omitempty"`
}

// PlatformConfig represents platform-specific configuration
type PlatformConfig struct {
	NumPciSegments int      `json:"num_pci_segments,omitempty"`
	IommuSegments  []int    `json:"iommu_segments,omitempty"`
	SerialNumber   string   `json:"serial_number,omitempty"`
	Uuid           string   `json:"uuid,omitempty"`
	OemStrings     []string `json:"oem_strings,omitempty"`
}

// NumaConfig represents NUMA configuration
type NumaConfig struct {
	GuestNumaId    int        `json:"guest_numa_id"`
	Cpus           []int      `json:"cpus,omitempty"`
	Distances      []NumaDist `json:"distances,omitempty"`
	MemoryZones    []string   `json:"memory_zones,omitempty"`
	SgxEpcSections []string   `json:"sgx_epc_sections,omitempty"`
}

// NumaDist represents NUMA distance
type NumaDist struct {
	Destination int `json:"destination"`
	Distance    int `json:"distance"`
}

// RateLimitGroup represents a rate limit group
type RateLimitGroup struct {
	Id          string       `json:"id"`
	RateLimiter *RateLimiter `json:"rate_limiter_config,omitempty"`
}

// RateLimiter represents rate limiting configuration
type RateLimiter struct {
	Bandwidth *TokenBucket `json:"bandwidth,omitempty"`
	Ops       *TokenBucket `json:"ops,omitempty"`
}

// TokenBucket represents a token bucket rate limiter
type TokenBucket struct {
	Size         uint64 `json:"size"`
	OneTimeBurst uint64 `json:"one_time_burst,omitempty"`
	RefillTime   uint64 `json:"refill_time"` // in milliseconds
}

// QueueAffinity represents queue affinity configuration
type QueueAffinity struct {
	Queue    int   `json:"queue"`
	HostCpus []int `json:"host_cpus"`
}

// VmInfo represents the response from the info endpoint
type VmInfo struct {
	Config           *VmConfig `json:"config,omitempty"`
	State            VmState   `json:"state"`
	MemoryActualSize uint64    `json:"memory_actual_size,omitempty"`
	DeviceTree       any       `json:"device_tree,omitempty"`
}

// SnapshotConfig represents snapshot configuration
type SnapshotConfig struct {
	DestinationUrl string `json:"destination_url"` // e.g., file:///path/to/snapshot
}

// RestoreConfig represents restore configuration
type RestoreConfig struct {
	SourceUrl string  `json:"source_url"`         // e.g., file:///path/to/snapshot
	Prefault  bool    `json:"prefault,omitempty"` // Prefault memory pages
	NetFds    []NetFd `json:"net_fds,omitempty"`  // Network FDs for restore
}

// NetFd represents network file descriptor mapping for restore
type NetFd struct {
	Id  string `json:"id"`
	Fds []int  `json:"fds"`
}

// ResizeConfig represents VM resize configuration
type ResizeConfig struct {
	DesiredVcpus   *int    `json:"desired_vcpus,omitempty"`
	DesiredRam     *uint64 `json:"desired_ram,omitempty"`
	DesiredBalloon *uint64 `json:"desired_balloon,omitempty"`
}

// ResizeDiskConfig represents disk resize configuration
type ResizeDiskConfig struct {
	DiskId  string `json:"disk_id"`
	NewSize uint64 `json:"new_size"` // in bytes
}

// MigrationConfig represents live migration configuration
type MigrationConfig struct {
	DestinationUrl string `json:"destination_url"` // tcp://host:port
}

// ReceiveMigrationConfig represents receive migration configuration
type ReceiveMigrationConfig struct {
	ReceiverUrl string `json:"receiver_url"` // tcp://0.0.0.0:port
}

// VmCounters represents VM performance counters
type VmCounters map[string]map[string]uint64

// SandboxInfo represents a sandbox instance managed by runner-ch
type SandboxInfo struct {
	Id           string            `json:"id"`
	State        VmState           `json:"state"`
	Vcpus        int               `json:"vcpus"`
	MemoryMB     uint64            `json:"memoryMB"`
	DiskPath     string            `json:"diskPath"`
	SnapshotPath string            `json:"snapshotPath,omitempty"`
	IpAddress    string            `json:"ipAddress,omitempty"`
	TapDevice    string            `json:"tapDevice,omitempty"`
	SocketPath   string            `json:"socketPath"`
	CreatedAt    time.Time         `json:"createdAt"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	GpuDevices   []string          `json:"gpuDevices,omitempty"`
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
