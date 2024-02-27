// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "github.com/daytonaio/daytona/pkg/os"

func GetBinaryUrls() map[os.OperatingSystem]string {
	return map[os.OperatingSystem]string{
		(os.Darwin_64_86):  "https://download.daytona.io/daytona/latest/daytona-darwin-amd64",
		(os.Darwin_arm64):  "https://download.daytona.io/daytona/latest/daytona-darwin-arm64",
		(os.Linux_64_86):   "https://download.daytona.io/daytona/latest/daytona-linux-amd64",
		(os.Linux_arm64):   "https://download.daytona.io/daytona/latest/daytona-linux-arm64",
		(os.Windows_64_86): "https://download.daytona.io/daytona/latest/daytona-windows-amd64.exe",
		(os.Windows_arm64): "https://download.daytona.io/daytona/latest/daytona-windows-arm64.exe",
	}
}

func GetIdeList() []Ide {
	return []Ide{
		{"vscode", "VS Code"},
		{"browser", "VS Code - Browser"},
	}
}

func GetGitProviderList() []GitProvider {
	return []GitProvider{
		{"github", "GitHub", ""},
		{"gitlab", "GitLab", ""},
		{"bitbucket", "Bitbucket", ""},
	}
}
