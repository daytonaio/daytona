// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/pkg/provider/manager"
	log "github.com/sirupsen/logrus"
)

func (s *Server) installDefaultProviders() error {
	manifest, err := s.ProviderManager.GetProvidersManifest()
	if err != nil {
		return err
	}

	defaultProviders := manifest.GetDefaultProviders()

	log.Info("Installing default providers")
	for providerName, provider := range defaultProviders {
		// Skip if default install lock file is present
		lockFilePath := filepath.Join(s.config.ProvidersDir, providerName, manager.DEFAULT_INSTALL_LOCK_FILE_NAME)
		_, err := os.Stat(lockFilePath)
		if err == nil {
			log.Infof("Skipping install for %s because it was manually removed", providerName)
			continue
		}

		_, err = s.ProviderManager.DownloadProvider(context.Background(), provider.DownloadUrls, providerName)
		if err != nil {
			if !manager.IsProviderAlreadyDownloaded(err, providerName) {
				log.Error(err)
			}
			continue
		}

		pluginPath, err := s.getPluginPath(filepath.Join(s.config.ProvidersDir, providerName))
		if err != nil {
			log.Error(err)
			continue
		}

		err = s.ProviderManager.RegisterProvider(pluginPath)
		if err != nil {
			log.Error(err)
			continue
		}

		err = s.ProviderManager.SetPresetTargets(providerName)
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Default providers installed")

	return nil
}

func (s *Server) registerProviders() error {
	log.Info("Registering providers")

	manifest, err := s.ProviderManager.GetProvidersManifest()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(s.config.ProvidersDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No providers found")
			return nil
		}
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := s.getPluginPath(filepath.Join(s.config.ProvidersDir, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			if strings.HasSuffix(pluginPath, manager.DEFAULT_INSTALL_LOCK_FILE_NAME) {
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

			if manifest.HasUpdateAvailable(info.Name, info.Version) {
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
