// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerUrl          string `envconfig:"SERVER_URL" validate:"required"`
	ApiToken           string `envconfig:"API_TOKEN" validate:"required"`
	ApiPort            int    `envconfig:"API_PORT"`
	TLSCertFile        string `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile         string `envconfig:"TLS_KEY_FILE"`
	EnableTLS          bool   `envconfig:"ENABLE_TLS"`
	CacheRetentionDays int    `envconfig:"CACHE_RETENTION_DAYS"`
	Environment        string `envconfig:"ENVIRONMENT"`
	ContainerRuntime   string `envconfig:"CONTAINER_RUNTIME"`
	ContainerNetwork   string `envconfig:"CONTAINER_NETWORK"`
	LogFilePath        string `envconfig:"LOG_FILE_PATH"`
	// S3 Configuration
	AWSRegion          string `envconfig:"S3_REGION"`
	AWSEndpointUrl     string `envconfig:"S3_ENDPOINT"`
	AWSAccessKeyId     string `envconfig:"S3_ACCESS_KEY"`
	AWSSecretAccessKey string `envconfig:"S3_SECRET_KEY"`
	AWSDefaultBucket   string `envconfig:"S3_DEFAULT_BUCKET"`
	// Sandbox Disk Configuration
	ResourceLimitsDisabled bool   `envconfig:"RESOURCE_LIMITS_DISABLED"`
	DataDir                string `envconfig:"DISK_DATA_DIR"`
	LayerSizeThresholdMB   int64  `envconfig:"DISK_LAYER_SIZE_THRESHOLD_MB"`
	Compression            string `envconfig:"DISK_COMPRESSION"`
	ClusterSize            int    `envconfig:"DISK_CLUSTER_SIZE"`
	LazyRefcounts          bool   `envconfig:"DISK_LAZY_REFCOUNTS"`
	Preallocation          string `envconfig:"DISK_PREALLOCATION"`
	MaxMounted             int    `envconfig:"DISK_MAX_MOUNTED"`
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

	if config.ApiPort == 0 {
		config.ApiPort = DEFAULT_API_PORT
	}

	if config.DataDir == "" {
		config.DataDir = filepath.Join(os.TempDir(), "daytona", "sdisk")
	}

	if config.LayerSizeThresholdMB == 0 {
		config.LayerSizeThresholdMB = 100
	}

	if config.Compression == "" {
		config.Compression = "zlib"
	}

	if config.ClusterSize == 0 {
		config.ClusterSize = 65536
	}

	if config.LazyRefcounts == false {
		config.LazyRefcounts = true
	}

	if config.Preallocation == "" {
		config.Preallocation = "off"
	}

	if config.MaxMounted == 0 {
		config.MaxMounted = 100
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
	buildId := snapshotRef
	if colonIndex := strings.Index(snapshotRef, ":"); colonIndex != -1 {
		buildId = snapshotRef[:colonIndex]
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
