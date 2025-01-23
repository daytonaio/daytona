// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/services"
	"golang.org/x/mod/semver"
)

type Version struct {
	DownloadUrls map[os.OperatingSystem]string `json:"downloadUrls"`
}

type ProvidersManifest map[string]ProviderManifest

type ProviderManifest struct {
	Default  bool               `json:"default"`
	Label    *string            `json:"label"`
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

func GetProviderDownloadUrls(name, version, registryUrl string) (map[os.OperatingSystem]string, error) {
	manifest, err := GetProvidersManifest(registryUrl)
	if err != nil {
		return nil, err
	}

	return (*manifest)[name].Versions[version].DownloadUrls, nil
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
		latestManifest[provider] = ProviderManifest{Default: manifest.Default, Label: manifest.Label, Versions: make(map[string]Version)}
		versionName, version := manifest.FindLatestVersion()
		latestManifest[provider].Versions[versionName] = *version
	}

	return &latestManifest
}

func (m *ProvidersManifest) GetProviderListFromManifest() []services.ProviderDTO {
	providerList := []services.ProviderDTO{}
	for providerName, providerManifest := range *m {
		latestVersion, _ := providerManifest.FindLatestVersion()
		for version := range providerManifest.Versions {
			providerList = append(providerList, services.ProviderDTO{
				Name:    providerName,
				Label:   providerManifest.Label,
				Version: version,
				Latest:  version == latestVersion,
			})
		}
	}

	slices.SortFunc(providerList, func(a, b services.ProviderDTO) int {
		return strings.Compare(a.Name, b.Name)
	})

	return providerList
}
