// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	api_util "github.com/daytonaio/daytona/pkg/server/api/util"
	log "github.com/sirupsen/logrus"
)

func (s *Server) downloadDefaultProviders() error {
	manifest, err := manager.GetProvidersManifest(s.Config.RegistryUrl)
	if err != nil {
		return err
	}

	defaultProviderPlugins := manager.GetDefaultProviders(*manifest)

	log.Info("Downloading default providers")
	for pluginName, plugin := range defaultProviderPlugins {
		downloadPath := filepath.Join(s.Config.ProvidersDir, pluginName, pluginName)
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

func (s *Server) registerProviders() error {
	log.Info("Registering providers")

	manifest, err := manager.GetProvidersManifest(s.Config.RegistryUrl)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(s.Config.ProvidersDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No providers found")
			return nil
		}
		return err
	}

	logsDir, err := GetWorkspaceLogsDir()
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := s.getPluginPath(filepath.Join(s.Config.ProvidersDir, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			err = manager.RegisterProvider(pluginPath, api_util.GetDaytonaScriptUrl(s.Config), util.GetFrpcServerUrl(s.Config), util.GetFrpcApiUrl(s.Config), logsDir)
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
				log.Infof("Update available for %s. Update with `daytona server provider update`.", info.Name)
			}
		}
	}

	log.Info("Providers registered")

	return nil
}

func (s *Server) getPluginPath(dir string) (string, error) {
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
