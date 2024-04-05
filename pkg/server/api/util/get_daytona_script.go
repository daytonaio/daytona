// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/url"

	"github.com/daytonaio/daytona/internal/util"
)

func GetDaytonaScriptUrl(protocol, serverId, serverDomain string) string {
	url, _ := url.JoinPath(util.GetFrpcApiUrl(protocol, serverId, serverDomain), "binary", "script")
	return url
}
