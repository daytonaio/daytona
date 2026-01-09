// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SetUpgraderSubprotocols(subprotocols []string) {
	if subprotocols == nil {
		subprotocols = []string{}
	}

	upgrader.Subprotocols = subprotocols
}

func IsCombinedOutput(sdkVersion string, versionComparison *int, requestHeader http.Header) bool {
	return (versionComparison != nil && *versionComparison < 0 && sdkVersion != "0.0.0-dev") || (sdkVersion == "" && requestHeader.Get("X-Daytona-Split-Output") != "true")
}
