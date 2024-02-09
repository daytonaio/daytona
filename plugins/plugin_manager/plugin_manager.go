package plugin_manager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	operating_system "github.com/daytonaio/daytona/common/os"
	"github.com/daytonaio/daytona/internal/util"
	"golang.org/x/mod/semver"
)

type PluginVersion struct {
	DownloadUrls map[operating_system.OperatingSystem]string `json:"downloadUrls"`
}

type PluginsManifest struct {
	ProvisionerPlugins  map[string]PluginManifest `json:"provisionerPlugins"`
	AgentServicePlugins map[string]PluginManifest `json:"agentServicePlugins"`
}

type PluginManifest struct {
	Default  bool                     `json:"default"`
	Versions map[string]PluginVersion `json:"versions"`
}

func GetPluginsManifest(registryUrl string) (*PluginsManifest, error) {
	manifestUrl := fmt.Sprintf("%s/manifest.json", registryUrl)

	resp, err := http.Get(manifestUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	manifestJson, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest PluginsManifest
	err = json.Unmarshal(manifestJson, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func DownloadPlugin(downloadUrls map[operating_system.OperatingSystem]string, downloadPath string) error {
	operatingSystem, err := util.GetOperatingSystem()
	if err != nil {
		return err
	}

	return downloadFile(downloadUrls[*operatingSystem], downloadPath)
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = os.MkdirAll(path.Dir(filepath), os.ModePerm)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func FindLatestVersion(pluginManifest PluginManifest) *PluginVersion {
	var latestVersion string = "v0.0.0"

	for version := range pluginManifest.Versions {
		if version == "latest" {
			continue
		}

		if semver.Compare(version, latestVersion) > 0 {
			latestVersion = version
		}
	}

	version, ok := pluginManifest.Versions[latestVersion]
	if !ok {
		return nil
	}

	return &version
}

func GetDefaultPlugins(plugins map[string]PluginManifest) map[string]*PluginVersion {
	defaultPlugins := make(map[string]*PluginVersion)
	for pluginName, pluginManifest := range plugins {
		if pluginManifest.Default {
			latestVersion, ok := pluginManifest.Versions["latest"]
			if !ok {
				latestVersion = *FindLatestVersion(pluginManifest)
			}
			defaultPlugins[pluginName] = &latestVersion
		}
	}

	return defaultPlugins
}
