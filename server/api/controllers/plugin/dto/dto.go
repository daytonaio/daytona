package dto

import "github.com/daytonaio/daytona/common/os"

type ProvisionerPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
} //	@name	ProvisionerPlugin

type AgentServicePlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
} //	@name	AgentServicePlugin

type InstallPluginRequest struct {
	Name         string                        `json:"name"`
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls"`
} //	@name	InstallPluginRequest
