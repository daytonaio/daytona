// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
)

func GetFrpcApiDomain(serverId, frpsDomain string) string {
	return fmt.Sprintf("api-%s", GetFrpcServerDomain(serverId, frpsDomain))
}

func GetFrpcServerDomain(serverId, frpsDomain string) string {
	return fmt.Sprintf("%s.%s", serverId, frpsDomain)
}

func GetFrpcHeadscaleUrl(protocol, serverId, frpsDomain string) string {
	return fmt.Sprintf("%s://%s", protocol, GetFrpcServerDomain(serverId, frpsDomain))
}

func GetFrpcApiUrl(protocol, serverId, frpsDomain string) string {
	return fmt.Sprintf("%s://%s", protocol, GetFrpcApiDomain(serverId, frpsDomain))
}

func GetFrpcRegistryDomain(serverId, frpsDomain string) string {
	return fmt.Sprintf("registry-%s", GetFrpcServerDomain(serverId, frpsDomain))
}

func GetFrpcRegistryUrl(protocol, serverId, frpsDomain string) string {
	return fmt.Sprintf("%s://%s", protocol, GetFrpcRegistryDomain(serverId, frpsDomain))
}
