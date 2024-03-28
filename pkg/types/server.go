// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package types

type FRPSConfig struct {
	Domain   string `json:"domain"`
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
} // @name FRPSConfig

type ServerConfig struct {
	ProvidersDir      string        `json:"providersDir"`
	RegistryUrl       string        `json:"registryUrl"`
	GitProviders      []GitProvider `json:"gitProviders"`
	Id                string        `json:"id"`
	ServerDownloadUrl string        `json:"serverDownloadUrl"`
	Frps              *FRPSConfig   `json:"frps,omitempty"`
	ApiPort           uint32        `json:"apiPort"`
	HeadscalePort     uint32        `json:"headscalePort"`
	TargetsFilePath   string        `json:"targetsFilePath"`
	BinariesPath      string        `json:"binariesPath"`
} // @name ServerConfig

type NetworkKey struct {
	Key string `json:"key"`
} // @name NetworkKey

type ApiKeyType string

const (
	ApiKeyTypeClient  ApiKeyType = "client"
	ApiKeyTypeProject ApiKeyType = "project"
)

type ApiKey struct {
	KeyHash string     `json:"keyHash"`
	Type    ApiKeyType `json:"type"`
	// Project or client name
	Name string `json:"name"`
} // @name ApiKey
