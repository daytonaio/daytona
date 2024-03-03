// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/pkg/os"
	"golang.org/x/mod/semver"
)

type Version struct {
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls"`
}

type ProvidersManifest map[string]ProviderManifest

type ProviderManifest struct {
	Default  bool               `json:"default"`
	Versions map[string]Version `json:"versions"`
}

func GetProvidersManifest(registryUrl string) (*ProvidersManifest, error) {
	manifestUrl := fmt.Sprintf("%s/providers/manifest.json", registryUrl)

	resp, err := http.Get(manifestUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	manifestJson, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest ProvidersManifest
	err = json.Unmarshal(manifestJson, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func DownloadProvider(downloadUrls map[os.OperatingSystem]string, downloadPath string) error {
	operatingSystem, err := os.GetOperatingSystem()
	if err != nil {
		return err
	}

	return os.DownloadFile(downloadUrls[*operatingSystem], downloadPath)
}

func FindLatestVersion(providerManifest ProviderManifest) *Version {
	var latestVersion string = "v0.0.0"

	for version := range providerManifest.Versions {
		if version == "latest" {
			continue
		}

		if semver.Compare(version, latestVersion) > 0 {
			latestVersion = version
		}
	}

	version, ok := providerManifest.Versions[latestVersion]
	if !ok {
		return nil
	}

	return &version
}

func GetDefaultProviders(manifest ProvidersManifest) map[string]*Version {
	defaultProviders := make(map[string]*Version)
	for providerName, providerManifest := range manifest {
		if providerManifest.Default {
			latestVersion, ok := providerManifest.Versions["latest"]
			if !ok {
				latestVersion = *FindLatestVersion(providerManifest)
			}
			defaultProviders[providerName] = &latestVersion
		}
	}

	return defaultProviders
}
