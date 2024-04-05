// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
)

type Self struct {
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
}

type TailscaleServer interface {
	Connect() error
	CreateAuthKey() (string, error)
	CreateUser() error
	HTTPClient() *http.Client
	Start() error
}

type FRPSConfig struct {
	Domain   string `json:"domain"`
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
} // @name FRPSConfig

type NetworkKey struct {
	Key string `json:"key"`
} // @name NetworkKey

type ServerConfig struct {
	ProvidersDir      string      `json:"providersDir"`
	RegistryUrl       string      `json:"registryUrl"`
	Id                string      `json:"id"`
	ServerDownloadUrl string      `json:"serverDownloadUrl"`
	Frps              *FRPSConfig `json:"frps,omitempty"`
	ApiPort           uint32      `json:"apiPort"`
	HeadscalePort     uint32      `json:"headscalePort"`
	TargetsFilePath   string      `json:"targetsFilePath"`
	BinariesPath      string      `json:"binariesPath"`
} // @name ServerConfig

type Server struct {
	Config ServerConfig

	TailscaleServer TailscaleServer
}
