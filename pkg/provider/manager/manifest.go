// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
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

func (p *ProviderManifest) FindLatestVersion() (string, *Version) {
	var latestVersion string = "v0.0.0"

	for version := range p.Versions {
		if version == "latest" {
			continue
		}

		if semver.Compare(version, latestVersion) > 0 {
			latestVersion = version
		}
	}

	version, ok := p.Versions[latestVersion]
	if !ok {
		return latestVersion, nil
	}

	return latestVersion, &version
}

func (p *ProvidersManifest) GetDefaultProviders() map[string]*Version {
	defaultProviders := make(map[string]*Version)
	for providerName, providerManifest := range *p {
		if providerManifest.Default {
			latestVersion, ok := providerManifest.Versions["latest"]
			if !ok {
				_, latest := providerManifest.FindLatestVersion()
				latestVersion = *latest
			}
			defaultProviders[providerName] = &latestVersion
		}
	}

	return defaultProviders
}

func (p *ProvidersManifest) HasUpdateAvailable(providerName string, currentVersion string) bool {
	provider, ok := (*p)[providerName]
	if !ok {
		return false
	}

	var latestVersion string = "v0.0.0"

	for version := range provider.Versions {
		if version == "latest" {
			continue
		}

		if semver.Compare(version, latestVersion) > 0 {
			latestVersion = version
		}
	}

	return semver.Compare(latestVersion, currentVersion) > 0
}

func (m *ProvidersManifest) GetLatestVersions() *ProvidersManifest {
	var latestManifest ProvidersManifest = make(map[string]ProviderManifest, 0)
	for provider, manifest := range *m {
		latestManifest[provider] = ProviderManifest{Default: manifest.Default, Versions: make(map[string]Version)}
		versionName, version := manifest.FindLatestVersion()
		latestManifest[provider].Versions[versionName] = *version
	}

	return &latestManifest
}
