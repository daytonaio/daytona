// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import "time"

// ContainerSummary represents a summary of a container
type ContainerSummary struct {
	ID      string
	Names   []string
	Image   string
	ImageID string
	Command string
	Created int64
	State   string
	Status  string
	Labels  map[string]string
}

// SystemInfo represents system-wide information
type SystemInfo struct {
	ID                 string
	Containers         int
	ContainersRunning  int
	ContainersPaused   int
	ContainersStopped  int
	Images             int
	Driver             string
	DriverStatus       [][2]string
	SystemTime         string
	LoggingDriver      string
	CgroupDriver       string
	CgroupVersion      string
	KernelVersion      string
	OperatingSystem    string
	OSType             string
	Architecture       string
	NCPU               int
	MemTotal           int64
	IndexServerAddress string
	RegistryConfig     interface{}
	GenericResources   []interface{}
	HTTPProxy          string
	HTTPSProxy         string
	NoProxy            string
	Name               string
	Labels             []string
	ExperimentalBuild  bool
	ServerVersion      string
	Runtimes           map[string]interface{}
	DefaultRuntime     string
	Swarm              interface{}
	LiveRestoreEnabled bool
	Isolation          string
	InitBinary         string
	ContainerdCommit   interface{}
	RuncCommit         interface{}
	InitCommit         interface{}
	SecurityOptions    []string
	Warnings           []string
}

// ContainerListOptions represents options for listing containers
type ContainerListOptions struct {
	All     bool
	Limit   int
	Size    bool
	Filters map[string][]string
}

// Resources represents container resource limits
type Resources struct {
	CPUShares            int64
	Memory               int64
	NanoCPUs             int64
	CgroupParent         string
	BlkioWeight          uint16
	BlkioWeightDevice    []interface{}
	BlkioDeviceReadBps   []interface{}
	BlkioDeviceWriteBps  []interface{}
	BlkioDeviceReadIOps  []interface{}
	BlkioDeviceWriteIOps []interface{}
	CPUPeriod            int64
	CPUQuota             int64
	CPURealtimePeriod    int64
	CPURealtimeRuntime   int64
	CpusetCpus           string
	CpusetMems           string
	Devices              []interface{}
	DeviceCgroupRules    []string
	DeviceRequests       []interface{}
	KernelMemory         int64
	KernelMemoryTCP      int64
	MemoryReservation    int64
	MemorySwap           int64
	MemorySwappiness     *int64
	OomKillDisable       *bool
	PidsLimit            *int64
	Ulimits              []interface{}
	CPUCount             int64
	CPUPercent           int64
	IOMaximumIOps        uint64
	IOMaximumBandwidth   uint64
}

// HostConfig represents host configuration for a container
type HostConfig struct {
	Binds           []string
	ContainerIDFile string
	LogConfig       interface{}
	NetworkMode     string
	PortBindings    map[string][]interface{}
	RestartPolicy   interface{}
	AutoRemove      bool
	VolumeDriver    string
	VolumesFrom     []string
	CapAdd          []string
	CapDrop         []string
	CgroupnsMode    string
	DNS             []string
	DNSOptions      []string
	DNSSearch       []string
	ExtraHosts      []string
	GroupAdd        []string
	IpcMode         string
	Cgroup          string
	Links           []string
	OomScoreAdj     int
	PidMode         string
	Privileged      bool
	PublishAllPorts bool
	ReadonlyRootfs  bool
	SecurityOpt     []string
	StorageOpt      map[string]string
	Tmpfs           map[string]string
	UTSMode         string
	UsernsMode      string
	ShmSize         int64
	Sysctls         map[string]string
	Runtime         string
	ConsoleSize     [2]uint
	Isolation       string
	Resources       Resources
	Mounts          []interface{}
	MaskedPaths     []string
	ReadonlyPaths   []string
	Init            *bool
}

// ContainerState represents the state of a container
type ContainerState struct {
	Status     string
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        int
	ExitCode   int
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
}

// ContainerConfig represents container configuration
type ContainerConfig struct {
	Hostname        string
	Domainname      string
	User            string
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	ExposedPorts    map[string]struct{}
	Tty             bool
	OpenStdin       bool
	StdinOnce       bool
	Env             []string
	Cmd             []string
	Healthcheck     interface{}
	ArgsEscaped     bool
	Image           string
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	MacAddress      string
	OnBuild         []string
	Labels          map[string]string
	StopSignal      string
	StopTimeout     *int
	Shell           []string
}

// NetworkSettings represents network settings for a container
type NetworkSettings struct {
	Bridge                 string
	SandboxID              string
	HairpinMode            bool
	LinkLocalIPv6Address   string
	LinkLocalIPv6PrefixLen int
	Ports                  map[string][]interface{}
	SandboxKey             string
	SecondaryIPAddresses   []interface{}
	SecondaryIPv6Addresses []interface{}
	EndpointID             string
	Gateway                string
	GlobalIPv6Address      string
	GlobalIPv6PrefixLen    int
	IPAddress              string
	IPPrefixLen            int
	IPv6Gateway            string
	MacAddress             string
	Networks               map[string]interface{}
}

// ContainerJSON represents detailed container information
type ContainerJSON struct {
	ID              string
	Created         time.Time
	Path            string
	Args            []string
	State           *ContainerState
	Image           string
	ResolvConfPath  string
	HostnamePath    string
	HostsPath       string
	LogPath         string
	Name            string
	RestartCount    int
	Driver          string
	Platform        string
	MountLabel      string
	ProcessLabel    string
	AppArmorProfile string
	ExecIDs         []string
	HostConfig      *HostConfig
	GraphDriver     interface{}
	Mounts          []interface{}
	Config          *ContainerConfig
	NetworkSettings *NetworkSettings
}
