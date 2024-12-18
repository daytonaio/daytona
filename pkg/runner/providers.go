// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
)

func (r *Runner) downloadDefaultProviders(registryUrl string) error {
	manifest, err := util.GetProvidersManifest(registryUrl)
	if err != nil {
		return err
	}

	defaultProviders := manifest.GetDefaultProviders()

	r.logger.Info("Downloading default providers")
	for providerName, provider := range defaultProviders {
		lockFilePath := filepath.Join(r.Config.ProvidersDir, providerName, providermanager.INITIAL_SETUP_LOCK_FILE_NAME)

		_, err := os.Stat(lockFilePath)
		if err == nil {
			continue
		}

		_, err = r.providerManager.DownloadProvider(context.Background(), provider.DownloadUrls, providerName)
		if err != nil {
			if !providermanager.IsProviderAlreadyDownloaded(err, providerName) {
				r.logger.Error(err)
			}
			continue
		}
	}

	r.logger.Info("Default providers downloaded")

	return nil
}

func (r *Runner) registerProviders(registryUrl string) error {
	r.logger.Info("Registering providers")

	manifest, err := util.GetProvidersManifest(registryUrl)
	if err != nil {
		return err
	}

	directoryEntries, err := os.ReadDir(r.Config.ProvidersDir)
	if err != nil {
		if os.IsNotExist(err) {
			r.logger.Info("No providers found")
			return nil
		}
		return err
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() {
			providerDir := filepath.Join(r.Config.ProvidersDir, entry.Name())

			pluginPath, err := r.getPluginPath(providerDir)
			if err != nil {
				if !providermanager.IsNoPluginFound(err, providerDir) {
					r.logger.Error(err)
				}
				continue
			}

			err = r.providerManager.RegisterProvider(pluginPath, false)
			if err != nil {
				r.logger.Error(err)
				continue
			}

			// Lock the initial setup
			lockFilePath := filepath.Join(r.Config.ProvidersDir, entry.Name(), providermanager.INITIAL_SETUP_LOCK_FILE_NAME)

			_, err = os.Stat(lockFilePath)
			if err != nil {
				file, err := os.Create(lockFilePath)
				if err != nil {
					return err
				}
				defer file.Close()
			}

			// Check for updates
			provider, err := r.providerManager.GetProvider(entry.Name())
			if err != nil {
				r.logger.Error(err)
				continue
			}

			info, err := (*provider).GetInfo()
			if err != nil {
				r.logger.Error(err)
				continue
			}
			requirements, err := (*provider).CheckRequirements()
			if err != nil {
				return err
			}
			for _, req := range *requirements {
				if req.Met {
					r.logger.Infof("Provider requirement met: %s", req.Reason)
				} else {
					r.logger.Warnf("Provider requirement not met: %s", req.Reason)
				}
			}

			if manifest.HasUpdateAvailable(info.Name, info.Version) {
				r.logger.Infof("Update available for %s. Update with `daytona provider update`.", info.Name)
			}
		}
	}

	r.logger.Info("Providers registered")

	return nil
}

func (r *Runner) getPluginPath(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() && file.Name() != providermanager.INITIAL_SETUP_LOCK_FILE_NAME {
			return filepath.Join(dir, file.Name()), nil
		}
	}

	return "", errors.New("no plugin found in " + dir)
}
