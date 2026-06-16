// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"net/http"

	semver "github.com/Masterminds/semver/v3"
)

const (
	sourceHeader = "X-Daytona-Source"
	goSDKSource  = "sdk-go"

	// First Go SDK release whose generated client tolerates unknown response
	// fields. Keep in sync with the release that ships that client.
	unknownFieldsTolerantMinSDKVersion = "0.188.0"
)

// ClientRejectsUnknownResponseFields reports whether the request comes from a Go
// SDK client old enough to decode responses with a strict JSON decoder, which
// breaks when the daemon adds new fields to a response type. Other sources, dev
// builds, and unparseable/absent versions default to tolerant (false).
func ClientRejectsUnknownResponseFields(header http.Header) bool {
	if header.Get(sourceHeader) != goSDKSource {
		return false
	}
	v := ExtractSdkVersionFromHeader(header)
	if v == "" {
		return false
	}
	if sv, err := semver.NewVersion(v); err == nil && sv.Major() == 0 && sv.Minor() == 0 && sv.Patch() == 0 {
		return false
	}
	cmp, err := CompareVersions(v, unknownFieldsTolerantMinSDKVersion)
	if err != nil {
		return false
	}
	return *cmp < 0
}
