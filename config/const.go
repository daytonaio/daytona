// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import "github.com/daytonaio/daytona/remote_installer"

func GetBinaryUrls() map[remote_installer.RemoteOS]string {
	return map[remote_installer.RemoteOS]string{
		(remote_installer.OSDarwin_64_86): "https://download.daytona.io/core/daytona-core-darwin-amd64",
		(remote_installer.OSDarwin_arm64): "https://download.daytona.io/core/daytona-core-darwin-arm64",
		(remote_installer.OSLinux_64_86):  "https://download.daytona.io/core/daytona-core-linux-amd64",
		(remote_installer.OSLinux_arm64):  "https://download.daytona.io/core/daytona-core-linux-arm64",
	}
}

func GetIdeList() []Ide {
	return []Ide{
		{"vscode", "VS Code"},
		{"browser", "VS Code - Browser"},
	}
}
