// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "fmt"

func GetProjectStartScript(daytonaDownloadUrl string, apiKey string) string {
	return fmt.Sprintf(`curl -sfL -H "Authorization: Bearer %s" %s | sudo -E bash && daytona agent`, apiKey, daytonaDownloadUrl)
}
