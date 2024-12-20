// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package providermanager

import "fmt"

func IsProviderAlreadyDownloaded(err error, name string) bool {
	return err.Error() == providerAlreadyDownloadedError(name).Error()
}

func IsNoPluginFound(err error, dir string) bool {
	return err.Error() == noPluginFoundError(dir).Error()
}

func providerAlreadyDownloadedError(name string) error {
	return fmt.Errorf("provider %s already installed", name)
}

func noPluginFoundError(dir string) error {
	return fmt.Errorf("no plugin found in %s", dir)
}
