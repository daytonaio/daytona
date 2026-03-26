// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"github.com/docker/cli/cli/config/configfile"
	clitypes "github.com/docker/cli/cli/config/types"
	docker_registry "github.com/docker/docker/api/types/registry"
)

func registryAuthConfigsToConfigFile(m map[string]docker_registry.AuthConfig) *configfile.ConfigFile {
	cf := &configfile.ConfigFile{AuthConfigs: make(map[string]clitypes.AuthConfig)}
	if m == nil {
		return cf
	}
	for k, v := range m {
		cf.AuthConfigs[k] = clitypes.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Auth:          v.Auth,
			ServerAddress: v.ServerAddress,
			IdentityToken: v.IdentityToken,
			RegistryToken: v.RegistryToken,
		}
	}
	return cf
}
