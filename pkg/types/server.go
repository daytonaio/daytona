// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package types

type FRPSConfig struct {
	Domain   string `json:"domain"`
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
} // @name FRPSConfig

type GitProvider struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
} // @name GitProvider

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
} // @name ServerConfig

type NetworkKey struct {
	Key string `json:"key"`
} // @name NetworkKey
