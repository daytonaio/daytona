// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/os"
)

type Provider struct {
	Name    string  `json:"name" validate:"required"`
	Label   *string `json:"label" validate:"optional"`
	Version string  `json:"version" validate:"required"`
} //	@name	Provider

type InstallProviderRequest struct {
	Name         string                        `json:"name" validate:"required"`
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls" validate:"required"`
} //	@name	InstallProviderRequest
