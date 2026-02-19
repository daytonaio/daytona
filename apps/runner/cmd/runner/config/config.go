// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/vishvananda/netlink"
)

type Config struct {
	DaytonaApiUrl                      string        `envconfig:"DAYTONA_API_URL"`
	ApiToken                           string        `envconfig:"DAYTONA_RUNNER_TOKEN"`
	ApiPort                            int           `envconfig:"API_PORT"`
	TLSCertFile                        string        `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile                         string        `envconfig:"TLS_KEY_FILE"`
	EnableTLS                          bool          `envconfig:"ENABLE_TLS"`
	CacheRetentionDays                 int           `envconfig:"CACHE_RETENTION_DAYS"`
	Environment                        string        `envconfig:"ENVIRONMENT"`
	ContainerRuntime                   string        `envconfig:"CONTAINER_RUNTIME"`
	ContainerNetwork                   string        `envconfig:"CONTAINER_NETWORK"`
	LogFilePath                        string        `envconfig:"LOG_FILE_PATH"`
	AWSRegion                          string        `envconfig:"AWS_REGION"`
	AWSEndpointUrl                     string        `envconfig:"AWS_ENDPOINT_URL"`
	AWSAccessKeyId                     string        `envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey                 string        `envconfig:"AWS_SECRET_ACCESS_KEY"`
	AWSDefaultBucket                   string        `envconfig:"AWS_DEFAULT_BUCKET"`
	ResourceLimitsDisabled             bool          `envconfig:"RESOURCE_LIMITS_DISABLED"`
	DaemonStartTimeoutSec              int           `envconfig:"DAEMON_START_TIMEOUT_SEC"`
	SandboxStartTimeoutSec             int           `envconfig:"SANDBOX_START_TIMEOUT_SEC"`
	UseSnapshotEntrypoint              bool          `envconfig:"USE_SNAPSHOT_ENTRYPOINT"`
	Domain                             string        `envconfig:"RUNNER_DOMAIN" validate:"omitempty,hostname|ip"`
	VolumeCleanupIntervalSec           int           `envconfig:"VOLUME_CLEANUP_INTERVAL_SEC" default:"30" validate:"min=10"`
	VolumeCleanupDryRun                bool          `envconfig:"VOLUME_CLEANUP_DRY_RUN" default:"true"`
	PollTimeout                        time.Duration `envconfig:"POLL_TIMEOUT" default:"30s"`
	PollLimit                          int           `envconfig:"POLL_LIMIT" default:"10" validate:"min=1,max=100"`
	CollectorWindowSize                int           `envconfig:"COLLECTOR_WINDOW_SIZE" default:"60" validate:"min=1"`
	CPUUsageSnapshotInterval           time.Duration `envconfig:"CPU_USAGE_SNAPSHOT_INTERVAL" default:"5s" validate:"min=1s"`
	AllocatedResourcesSnapshotInterval time.Duration `envconfig:"ALLOCATED_RESOURCES_SNAPSHOT_INTERVAL" default:"5s" validate:"min=1s"`
	HealthcheckInterval                time.Duration `envconfig:"HEALTHCHECK_INTERVAL" default:"30s" validate:"min=10s"`
	HealthcheckTimeout                 time.Duration `envconfig:"HEALTHCHECK_TIMEOUT" default:"10s"`
	BackupTimeoutMin                   int           `envconfig:"BACKUP_TIMEOUT_MIN" default:"60" validate:"min=1"`
	ApiVersion                         int           `envconfig:"API_VERSION" default:"2"`
	InitializeDaemonTelemetry          bool          `envconfig:"INITIALIZE_DAEMON_TELEMETRY" default:"true"`
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

	if config.DaytonaApiUrl == "" {
		// For backward compatibility
		serverUrl := os.Getenv("SERVER_URL")
		if serverUrl == "" {
			return nil, fmt.Errorf("DAYTONA_API_URL or SERVER_URL is required")
		}
		config.DaytonaApiUrl = serverUrl
	}

	if config.ApiToken == "" {
		// For backward compatibility
		apiToken := os.Getenv("API_TOKEN")
		if apiToken == "" {
			return nil, fmt.Errorf("DAYTONA_RUNNER_TOKEN or API_TOKEN is required")
		}
		config.ApiToken = apiToken
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

func GetContainerRuntime() string {
	return config.ContainerRuntime
}

func GetContainerNetwork() string {
	return config.ContainerNetwork
}

func GetEnvironment() string {
	return config.Environment
}

func GetBuildLogFilePath(snapshotRef string) (string, error) {
	// Extract image name from various snapshot ref formats:
	// - registry:5000/daytona/daytona-<hash>
	// - daytona-<hash>
	// - daytona-<hash>:tag
	// - cr.preprod.daytona.io/sbox/daytona/daytona-<hash>:daytona

	buildId := snapshotRef

	// Remove tag if present (everything after last colon that's not part of a port)
	// A tag colon will come after the last slash
	lastSlashIndex := strings.LastIndex(buildId, "/")
	lastColonIndex := strings.LastIndex(buildId, ":")

	if lastColonIndex > lastSlashIndex && lastColonIndex != -1 {
		// This colon is a tag separator, not a port separator
		buildId = buildId[:lastColonIndex]
	}

	// Extract the image name (last component after the last slash)
	if lastSlashIndex := strings.LastIndex(buildId, "/"); lastSlashIndex != -1 {
		buildId = buildId[lastSlashIndex+1:]
	}

	c, err := GetConfig()
	if err != nil {
		return "", err
	}

	logPath := filepath.Join(filepath.Dir(c.LogFilePath), "builds", buildId)

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	if _, err := os.OpenFile(logPath, os.O_CREATE, 0644); err != nil {
		return "", fmt.Errorf("failed to create log file: %w", err)
	}

	return logPath, nil
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
