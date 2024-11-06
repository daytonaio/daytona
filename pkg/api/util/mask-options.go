// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"

	"github.com/daytonaio/daytona/pkg/server"
)

func GetMaskedOptions(server *server.Server, providerName, options string) (string, error) {
	p, err := server.ProviderManager.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	manifest, err := (*p).GetTargetConfigManifest()
	if err != nil {
		return "", err
	}

	var opts map[string]interface{}
	err = json.Unmarshal([]byte(options), &opts)
	if err != nil {
		return "", err
	}

	for name, property := range *manifest {
		if property.InputMasked {
			delete(opts, name)
		}
	}

	updatedOptions, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return "", err
	}

	return string(updatedOptions), nil
}
