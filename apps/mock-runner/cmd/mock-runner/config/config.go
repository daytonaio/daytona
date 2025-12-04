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
	LogFilePath        string `envconfig:"LOG_FILE_PATH"`
	ToolboxImage       string `envconfig:"TOOLBOX_IMAGE"`
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

	if config.ToolboxImage == "" {
		config.ToolboxImage = "ubuntu:22.04"
	}

	return config, nil
}

func GetEnvironment() string {
	if config == nil {
		return os.Getenv("ENVIRONMENT")
	}
	return config.Environment
}

func GetToolboxImage() string {
	if config == nil {
		img := os.Getenv("TOOLBOX_IMAGE")
		if img == "" {
			return "ubuntu:22.04"
		}
		return img
	}
	return config.ToolboxImage
}

func GetBuildLogFilePath(snapshotRef string) (string, error) {
	buildId := snapshotRef
	if colonIndex := strings.Index(snapshotRef, ":"); colonIndex != -1 {
		buildId = snapshotRef[:colonIndex]
	}

	c, err := GetConfig()
	if err != nil {
		// Fallback to temp directory
		logPath := filepath.Join(os.TempDir(), "mock-runner", "builds", buildId)
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			return "", fmt.Errorf("failed to create log directory: %w", err)
		}
		return logPath, nil
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
