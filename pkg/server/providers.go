// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/provider/manager"
	log "github.com/sirupsen/logrus"
)

func (s *Server) downloadDefaultProviders() error {
	manifest, err := s.ProviderManager.GetProvidersManifest()
	if err != nil {
		return err
	}

	defaultProviders := manager.GetDefaultProviders(*manifest)

	log.Info("Downloading default providers")
	for providerName, provider := range defaultProviders {
		_, err = s.ProviderManager.DownloadProvider(provider.DownloadUrls, providerName, false)
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Default providers downloaded")

	return nil
}

func (s *Server) registerProviders() error {
	log.Info("Registering providers")

	manifest, err := s.ProviderManager.GetProvidersManifest()
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

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := s.getPluginPath(filepath.Join(s.Config.ProvidersDir, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			err = s.ProviderManager.RegisterProvider(pluginPath)
			if err != nil {
				log.Error(err)
				continue
			}

			// Check for updates
			provider, err := s.ProviderManager.GetProvider(file.Name())
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
