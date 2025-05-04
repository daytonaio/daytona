// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ApiToken           string `envconfig:"API_TOKEN" validate:"required"`
	DaytonaBinaryUrl   string `envconfig:"DAYTONA_BINARY_URL" validate:"required"`
	TerminalBinaryUrl  string `envconfig:"TERMINAL_BINARY_URL" validate:"required"`
	ApiPort            int    `envconfig:"API_PORT"`
	TLSCertFile        string `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile         string `envconfig:"TLS_KEY_FILE"`
	EnableTLS          bool   `envconfig:"ENABLE_TLS"`
	DaytonaBinaryPath  string `envconfig:"DAYTONA_BINARY_PATH"`
	TerminalBinaryPath string `envconfig:"TERMINAL_BINARY_PATH"`
	CacheRetentionDays int    `envconfig:"CACHE_RETENTION_DAYS"`
	NodeEnv            string `envconfig:"NODE_ENV"`
	ContainerRuntime   string `envconfig:"CONTAINER_RUNTIME"`
	LogFilePath        string `envconfig:"LOG_FILE_PATH"`
	AWSRegion          string `envconfig:"AWS_REGION"`
	AWSEndpointUrl     string `envconfig:"AWS_ENDPOINT_URL"`
	AWSAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY"`
}

var DEFAULT_API_PORT int = 8080

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

	if config.ApiPort == 0 {
		config.ApiPort = DEFAULT_API_PORT
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if config.DaytonaBinaryPath == "" {
		config.DaytonaBinaryPath = path.Join(pwd, ".tmp", "binaries", "daytona")
	}

	if config.TerminalBinaryPath == "" {
		config.TerminalBinaryPath = path.Join(pwd, ".tmp", "binaries", "terminal")
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
