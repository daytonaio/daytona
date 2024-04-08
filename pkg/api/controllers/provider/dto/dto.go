// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/os"
)

type Provider struct {
	Name    string `json:"name"`
	Version string `json:"version"`
} //	@name	Provider

type InstallProviderRequest struct {
	Name         string                        `json:"name"`
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls"`
} //	@name	InstallProviderRequest
