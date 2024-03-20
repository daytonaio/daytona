// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/types"
	log "github.com/sirupsen/logrus"
)

func downloadDefaultProviders() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	manifest, err := manager.GetProvidersManifest(c.RegistryUrl)
	if err != nil {
		return err
	}

	defaultProviderPlugins := manager.GetDefaultProviders(*manifest)

	log.Info("Downloading default providers")
	for pluginName, plugin := range defaultProviderPlugins {
		downloadPath := filepath.Join(c.ProvidersDir, pluginName, pluginName)
		if runtime.GOOS == "windows" {
			downloadPath += ".exe"
		}

		if _, err := os.Stat(downloadPath); err == nil {
			log.Info(pluginName + " already downloaded")
			continue
		}
		log.Info("Downloading " + pluginName)
		err = manager.DownloadProvider(plugin.DownloadUrls, downloadPath)
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Default providers downloaded")

	return nil
}

func registerProviders(c *types.ServerConfig) error {
	log.Info("Registering providers")

	manifest, err := manager.GetProvidersManifest(c.RegistryUrl)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(c.ProvidersDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No providers found")
			return nil
		}
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := getPluginPath(filepath.Join(c.ProvidersDir, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			err = manager.RegisterProvider(pluginPath, c.ServerDownloadUrl, frpc.GetServerUrl(c), frpc.GetApiUrl(c))
			if err != nil {
				log.Error(err)
				continue
			}

			// Check for updates
			provider, err := manager.GetProvider(file.Name())
			if err != nil {
				log.Error(err)
				continue
			}

			info, err := (*provider).GetInfo()
			if err != nil {
				log.Error(err)
				continue
			}

			if manager.HasUpdateAvailable(info.Name, info.Version, *manifest) {
				log.Infof("Update available for %s. Update with `daytona provider update`.", info.Name)
			}
		}
	}

	log.Info("Providers registered")

	return nil
}

func getPluginPath(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			return filepath.Join(dir, file.Name()), nil
		}
	}

	return "", errors.New("no plugin found in " + dir)
}
