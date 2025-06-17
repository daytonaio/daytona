// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Token              string `envconfig:"TOKEN" validate:"required"`
	Port               int    `envconfig:"PORT"`
	TLSCertFile        string `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile         string `envconfig:"TLS_KEY_FILE"`
	EnableTLS          bool   `envconfig:"ENABLE_TLS"`
	CacheRetentionDays int    `envconfig:"CACHE_RETENTION_DAYS"`
	NodeEnv            string `envconfig:"NODE_ENV"`
	ContainerRuntime   string `envconfig:"CONTAINER_RUNTIME"`
	LogFilePath        string `envconfig:"LOG_FILE_PATH"`
	AWSRegion          string `envconfig:"AWS_REGION"`
	AWSEndpointUrl     string `envconfig:"AWS_ENDPOINT_URL"`
	AWSAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY"`
	AWSDefaultBucket   string `envconfig:"AWS_DEFAULT_BUCKET"`
	MetricsPort        int    `envconfig:"METRICS_PORT"`
	ProxyPort          int    `envconfig:"PROXY_PORT"`
}

var DEFAULT_PORT int = 3003
var DEFAULT_METRICS_PORT int = 9090
var DEFAULT_PROXY_PORT int = 3004
var config *Config

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
		// Continue anyway, as environment variables might be set directly
	}

	err = envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	var validate = validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	if config.Port == 0 {
		config.Port = DEFAULT_PORT
	}

	if config.MetricsPort == 0 {
		config.MetricsPort = DEFAULT_METRICS_PORT
	}

	if config.ProxyPort == 0 {
		config.ProxyPort = DEFAULT_PROXY_PORT
	}

	return config, nil
}

func GetContainerRuntime() string {
	return config.ContainerRuntime
}

func GetNodeEnv() string {
	return config.NodeEnv
}

func GetBuildLogFilePath(imageRef string) (string, error) {
	buildId := imageRef
	if colonIndex := strings.Index(imageRef, ":"); colonIndex != -1 {
		buildId = imageRef[:colonIndex]
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
