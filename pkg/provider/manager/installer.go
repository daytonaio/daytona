// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	goos "os"
	"path/filepath"
	"runtime"

	"github.com/daytonaio/daytona/pkg/os"
	log "github.com/sirupsen/logrus"
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

func (m *ProviderManager) GetProvidersManifest() (*ProvidersManifest, error) {
	manifestUrl := fmt.Sprintf("%s/providers/manifest.json", m.registryUrl)

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

func (m *ProviderManager) DownloadProvider(downloadUrls map[os.OperatingSystem]string, providerName string, throwIfPresent bool) (string, error) {
	downloadPath := filepath.Join(m.baseDir, providerName, providerName)
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	if _, err := goos.Stat(downloadPath); err == nil {
		if throwIfPresent {
			return "", fmt.Errorf("provider %s already downloaded", providerName)
		}
		return "", nil
	}

	log.Info("Downloading " + providerName)

	operatingSystem, err := os.GetOperatingSystem()
	if err != nil {
		return "", err
	}

	err = os.DownloadFile(downloadUrls[*operatingSystem], downloadPath)
	if err != nil {
		return "", err
	}

	return downloadPath, nil
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

func HasUpdateAvailable(providerName string, currentVersion string, manifest ProvidersManifest) bool {
	provider, ok := manifest[providerName]
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

// FIXME: temporary pollyfill
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
