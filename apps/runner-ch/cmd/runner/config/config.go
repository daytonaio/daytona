// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"fmt"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/vishvananda/netlink"
)

type Config struct {
	// Daytona API configuration
	ServerUrl string `envconfig:"SERVER_URL"`
	ApiToken  string `envconfig:"API_TOKEN"`

	// Runner API configuration
	ApiPort     int    `envconfig:"API_PORT"`
	TLSCertFile string `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile  string `envconfig:"TLS_KEY_FILE"`
	EnableTLS   bool   `envconfig:"ENABLE_TLS"`

	// Environment
	Environment        string `envconfig:"ENVIRONMENT"`
	CacheRetentionDays int    `envconfig:"CACHE_RETENTION_DAYS"`
	LogFilePath        string `envconfig:"LOG_FILE_PATH"`

	// Cloud Hypervisor configuration
	CHSocketsPath     string `envconfig:"CH_SOCKETS_PATH" default:"/var/run/cloud-hypervisor"`
	CHSandboxesPath   string `envconfig:"CH_SANDBOXES_PATH" default:"/var/lib/cloud-hypervisor/sandboxes"`
	CHSnapshotsPath   string `envconfig:"CH_SNAPSHOTS_PATH" default:"/var/lib/cloud-hypervisor/snapshots"`
	CHKernelPath      string `envconfig:"CH_KERNEL_PATH" default:"/var/lib/cloud-hypervisor/kernels/vmlinuz-6.8.0-90-generic"`
	CHInitramfsPath   string `envconfig:"CH_INITRAMFS_PATH" default:"/var/lib/cloud-hypervisor/kernels/initrd.img-6.8.0-90-generic"`
	CHFirmwarePath    string `envconfig:"CH_FIRMWARE_PATH" default:"/var/lib/cloud-hypervisor/firmware/hypervisor-fw"`
	CHBaseImagePath   string `envconfig:"CH_BASE_IMAGE_PATH" default:"/var/lib/cloud-hypervisor/snapshots/ubuntu-base.1/disk.qcow2"`
	CHBridgeName      string `envconfig:"CH_BRIDGE_NAME" default:"br0"`
	CHTapCreateScript string `envconfig:"CH_TAP_CREATE_SCRIPT" default:"/usr/local/bin/ch-create-tap"`
	CHTapDeleteScript string `envconfig:"CH_TAP_DELETE_SCRIPT" default:"/usr/local/bin/ch-delete-tap"`

	// Remote CH host configuration (empty for local mode)
	CHSSHHost    string `envconfig:"CH_SSH_HOST"`
	CHSSHKeyPath string `envconfig:"CH_SSH_KEY_PATH"`

	// Default VM resources
	CHDefaultCpus     int `envconfig:"CH_DEFAULT_CPUS" default:"2"`
	CHDefaultMemoryMB int `envconfig:"CH_DEFAULT_MEMORY_MB" default:"2048"`
	CHDefaultDiskGB   int `envconfig:"CH_DEFAULT_DISK_GB" default:"20"`

	// TAP pool configuration (pre-created TAP interfaces for fast VM creation)
	TapPoolEnabled bool `envconfig:"TAP_POOL_ENABLED" default:"true"`
	TapPoolSize    int  `envconfig:"TAP_POOL_SIZE" default:"10"`

	// VM pool configuration (for fast fork)
	VMPoolEnabled bool `envconfig:"VMPOOL_ENABLED" default:"false"`
	VMPoolSize    int  `envconfig:"VMPOOL_SIZE" default:"5"`

	// GPU passthrough
	GPUPassthroughEnabled bool `envconfig:"GPU_PASSTHROUGH_ENABLED" default:"false"`

	// Object storage (S3) configuration
	AWSRegion          string `envconfig:"AWS_REGION"`
	AWSEndpointUrl     string `envconfig:"AWS_ENDPOINT_URL"`
	AWSAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY"`
	AWSDefaultBucket   string `envconfig:"AWS_DEFAULT_BUCKET"`

	// Timeouts
	DaemonStartTimeoutSec  int `envconfig:"DAEMON_START_TIMEOUT_SEC" default:"60"`
	SandboxStartTimeoutSec int `envconfig:"SANDBOX_START_TIMEOUT_SEC" default:"30"`

	// SSH Gateway
	SSHGatewayEnabled bool `envconfig:"SSH_GATEWAY_ENABLE" default:"false"`

	// Runner identification
	Domain string `envconfig:"RUNNER_DOMAIN" validate:"omitempty,hostname|ip"`

	// Polling configuration (v2 API)
	ApiVersion          int           `envconfig:"API_VERSION" default:"2"`
	PollTimeout         time.Duration `envconfig:"POLL_TIMEOUT" default:"30s"`
	PollLimit           int           `envconfig:"POLL_LIMIT" default:"10" validate:"min=1,max=100"`
	HealthcheckInterval time.Duration `envconfig:"HEALTHCHECK_INTERVAL" default:"30s" validate:"min=10s"`
	HealthcheckTimeout  time.Duration `envconfig:"HEALTHCHECK_TIMEOUT" default:"10s"`

	// Memory ballooning configuration
	MemoryBallooningEnabled     bool    `envconfig:"MEMORY_BALLOONING_ENABLED" default:"true"`
	MemoryBallooningIntervalSec int     `envconfig:"MEMORY_BALLOONING_INTERVAL_SEC" default:"30"`
	MemoryBallooningMinVMGB     int     `envconfig:"MEMORY_BALLOONING_MIN_VM_GB" default:"4"`
	MemoryBallooningBufferGB    int     `envconfig:"MEMORY_BALLOONING_BUFFER_GB" default:"2"`
	MemoryBallooningBufferRatio float64 `envconfig:"MEMORY_BALLOONING_BUFFER_RATIO" default:"0.25"`
}

var DEFAULT_API_PORT int = 8080

var config *Config

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	var validate = validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	if config.ServerUrl == "" {
		return nil, fmt.Errorf("SERVER_URL is required")
	}

	if config.ApiToken == "" {
		return nil, fmt.Errorf("API_TOKEN is required")
	}

	if config.ApiPort == 0 {
		config.ApiPort = DEFAULT_API_PORT
	}

	if config.Domain == "" {
		ip, err := getOutboundIP()
		if err != nil {
			return nil, err
		}
		config.Domain = ip.String()
	}

	return config, nil
}

func GetEnvironment() string {
	if config == nil {
		return ""
	}
	return config.Environment
}

// getOutboundIP returns the IP address of the default route's network interface
func getOutboundIP() (net.IP, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}

	// Find the default route (destination 0.0.0.0/0)
	for _, route := range routes {
		if route.Dst == nil || route.Dst.IP.Equal(net.IPv4zero) {
			// Get the link (interface) for this route
			link, err := netlink.LinkByIndex(route.LinkIndex)
			if err != nil {
				return nil, fmt.Errorf("failed to get link: %w", err)
			}

			// Get addresses for this interface
			addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
			if err != nil {
				return nil, fmt.Errorf("failed to get addresses: %w", err)
			}

			if len(addrs) > 0 {
				return addrs[0].IP, nil
			}
		}
	}

	return nil, fmt.Errorf("no default route found")
}
