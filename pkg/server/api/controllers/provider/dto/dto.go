package dto

import (
	"github.com/daytonaio/daytona/pkg/os"
)

type Provider struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Targets []TargetDTO `json:"targets"`
} //	@name	Provider

type TargetDTO struct {
	Name string
	// JSON encoded map of options
	Options string
} //	@name	TargetDTO

type InstallProviderRequest struct {
	Name         string                        `json:"name"`
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls"`
} //	@name	InstallProviderRequest
