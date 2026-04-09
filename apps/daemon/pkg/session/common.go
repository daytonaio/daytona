// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"net/http"
	"strings"

	"github.com/daytonaio/daemon/internal/util"
)

func isDevVersion(version string) bool {
	return strings.Contains(version, "dev")
}

func IsCombinedOutput(sdkVersion string, versionComparison *int, requestHeader http.Header) bool {
	return (versionComparison != nil && *versionComparison < 0 && !isDevVersion(sdkVersion)) || (sdkVersion == "" && requestHeader.Get("X-Daytona-Split-Output") != "true")
}

func SkipServerDemux(sdkVersion string) bool {
	if sdkVersion == "" || isDevVersion(sdkVersion) {
		return false
	}
	comparison, err := util.CompareVersions(sdkVersion, "0.163.0-0")
	if err != nil {
		return false
	}
	return comparison != nil && *comparison < 0
}
