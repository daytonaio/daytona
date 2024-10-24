// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import "fmt"

func providerAlreadyDownloadedError(name string) error {
	return fmt.Errorf("provider %s already installed", name)
}

func IsProviderAlreadyDownloaded(err error, name string) bool {
	return err.Error() == providerAlreadyDownloadedError(name).Error()
}
