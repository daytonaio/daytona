// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net"
	"net/http"
)

type TailscaleServer interface {
	Connect() error
	CreateAuthKey() (string, error)
	CreateUser() error
	HTTPClient() *http.Client
	Dial(ctx context.Context, network, address string) (net.Conn, error)
	Start(errChan chan error) error
	Stop() error
	Purge() error
}

type ILocalContainerRegistry interface {
	Start() error
	Stop() error
	Purge() error
}

type FRPSConfig struct {
	Domain   string `json:"domain" validate:"required"`
	Port     uint32 `json:"port" validate:"required"`
	Protocol string `json:"protocol" validate:"required"`
} // @name FRPSConfig

type NetworkKey struct {
	Key string `json:"key" validate:"required"`
} // @name NetworkKey

type Config struct {
	ProvidersDir              string         `json:"providersDir" validate:"required"`
	RegistryUrl               string         `json:"registryUrl" validate:"required"`
	Id                        string         `json:"id" validate:"required"`
	ServerDownloadUrl         string         `json:"serverDownloadUrl" validate:"required"`
	Frps                      *FRPSConfig    `json:"frps,omitempty" validate:"optional"`
	ApiPort                   uint32         `json:"apiPort" validate:"required"`
	HeadscalePort             uint32         `json:"headscalePort" validate:"required"`
	BinariesPath              string         `json:"binariesPath" validate:"required"`
	LogFile                   *LogFileConfig `json:"logFile" validate:"required"`
	DefaultProjectImage       string         `json:"defaultProjectImage" validate:"required"`
	DefaultProjectUser        string         `json:"defaultProjectUser" validate:"required"`
	BuilderImage              string         `json:"builderImage" validate:"required"`
	LocalBuilderRegistryPort  uint32         `json:"localBuilderRegistryPort" validate:"required"`
	LocalBuilderRegistryImage string         `json:"localBuilderRegistryImage" validate:"required"`
	BuilderRegistryServer     string         `json:"builderRegistryServer" validate:"required"`
	BuildImageNamespace       string         `json:"buildImageNamespace" validate:"optional"`
	SamplesIndexUrl           string         `json:"samplesIndexUrl" validate:"optional"`
} // @name ServerConfig

type LogFileConfig struct {
	Path       string `json:"path" validate:"required"`
	MaxSize    int    `json:"maxSize" validate:"required"`
	MaxBackups int    `json:"maxBackups" validate:"required"`
	MaxAge     int    `json:"maxAge" validate:"required"`
	LocalTime  bool   `json:"localTime" validate:"optional"`
	Compress   bool   `json:"compress" validate:"optional"`
} // @name LogFileConfig
