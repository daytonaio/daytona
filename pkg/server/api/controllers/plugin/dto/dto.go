package dto

import "github.com/daytonaio/daytona/pkg/os"

type ProviderPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
} //	@name	ProviderPlugin

type AgentServicePlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
} //	@name	AgentServicePlugin

type InstallPluginRequest struct {
	Name         string                        `json:"name"`
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls"`
} //	@name	InstallPluginRequest
