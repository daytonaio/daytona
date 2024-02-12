// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/daytonaio/daytona/common/os"
)

func GetBinaryUrls() map[os.OperatingSystem]string {
	return map[os.OperatingSystem]string{
		(os.Darwin_64_86): "https://download.daytona.io/core/daytona-core-darwin-amd64",
		(os.Darwin_arm64): "https://download.daytona.io/core/daytona-core-darwin-arm64",
		(os.Linux_64_86):  "https://download.daytona.io/core/daytona-core-linux-amd64",
		(os.Linux_arm64):  "https://download.daytona.io/core/daytona-core-linux-arm64",
	}
}

func GetIdeList() []Ide {
	return []Ide{
		{"vscode", "VS Code"},
		{"browser", "VS Code - Browser"},
	}
}
