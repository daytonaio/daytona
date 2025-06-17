// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/base64"
	"encoding/json"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/docker/docker/api/types/registry"
)

func GetRegistryAuth(reg *pb.Registry) string {
	if reg == nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	authConfig := registry.AuthConfig{
		// TODO: nil-check
		Username: *reg.Username,
		Password: *reg.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		// Sometimes registry auth fails if "" is sent, so sending "empty" instead
		return "empty"
	}

	return base64.URLEncoding.EncodeToString(encodedJSON)
}
