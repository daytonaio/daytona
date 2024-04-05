// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/types"
)

func GetFrpcApiDomain(c types.ServerConfig) string {
	return fmt.Sprintf("api-%s", GetFrpcServerDomain(c))
}

func GetFrpcServerDomain(c types.ServerConfig) string {
	return fmt.Sprintf("%s.%s", c.Id, c.Frps.Domain)
}

func GetFrpcServerUrl(c types.ServerConfig) string {
	return fmt.Sprintf("%s://%s", c.Frps.Protocol, GetFrpcServerDomain(c))
}

func GetFrpcApiUrl(c types.ServerConfig) string {
	return fmt.Sprintf("%s://%s", c.Frps.Protocol, GetFrpcApiDomain(c))
}
