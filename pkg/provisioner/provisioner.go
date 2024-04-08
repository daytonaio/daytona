// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import "github.com/daytonaio/daytona/pkg/provider/manager"

type ProvisionerConfig struct {
	ProviderManager manager.ProviderManager
}

func NewProvisioner(config ProvisionerConfig) *Provisioner {
	return &Provisioner{
		providerManager: config.ProviderManager,
	}
}

type Provisioner struct {
	providerManager manager.ProviderManager
}
