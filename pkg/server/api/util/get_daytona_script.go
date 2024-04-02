// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/url"

	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/types"
)

func GetDaytonaScriptUrl(c *types.ServerConfig) string {
	url, _ := url.JoinPath(frpc.GetApiUrl(c), "binary", "script")
	return url
}
