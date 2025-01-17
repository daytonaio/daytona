// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "github.com/daytonaio/daytona/pkg/views"

func NotifyEmptyProviderList(tip bool) {
	views.RenderInfoMessageBold("No providers found")
	if tip {
		views.RenderTip("Use 'daytona provider install' to install a provider")
	}
}

func NotifyEmptyGitProviderList(tip bool) {
	views.RenderInfoMessageBold("No Git providers found")
	if tip {
		views.RenderTip("Use 'daytona git-provider create' to add a Git provider")
	}
}

func NotifyEmptyTargetConfigList(tip bool) {
	views.RenderInfoMessageBold("No target configs found")
	if tip {
		views.RenderTip("Use 'daytona target-config create' to create a target config")
	}
}

func NotifyEmptyWorkspaceTemplateList(tip bool) {
	views.RenderInfoMessageBold("No workspace templates found")
	if tip {
		views.RenderTip("Use 'daytona template create' to add a workspace template")
	}
}

func NotifyEmptyTargetList(tip bool) {
	views.RenderInfoMessageBold("No targets found")
	if tip {
		views.RenderTip("Use 'daytona create' to create a target")
	}
}

func NotifyEmptyWorkspaceList(tip bool) {
	views.RenderInfoMessageBold("No workspaces found")
	if tip {
		views.RenderTip("Use 'daytona create' to create a workspace")
	}
}

func NotifyEmptyProfileList(tip bool) {
	views.RenderInfoMessageBold("No profiles found")
	if tip {
		views.RenderTip("Use 'daytona profile create' to add a profile")
	}
}

func NotifyEmptyPrebuildList(tip bool) {
	views.RenderInfoMessageBold("No prebuilds found")
	if tip {
		views.RenderTip("Use 'daytona prebuild create' to add a prebuild")
	}
}

func NotifyEmptyApiKeyList(tip bool) {
	views.RenderInfoMessageBold("No API keys found")
	if tip {
		views.RenderTip("Use 'daytona api-key create' to create an API key")
	}
}

func NotifyEmptyBuildList(tip bool) {
	views.RenderInfoMessageBold("No builds found")
	if tip {
		views.RenderTip("Use 'daytona build run' to run a build or 'daytona prebuild create' to configure a prebuild rule")
	}
}

func NotifyEmptyEnvVarList(tip bool) {
	views.RenderInfoMessageBold("No server environment variables found")
	if tip {
		views.RenderTip("Use 'daytona env set' to add new server environment variables")
	}
}

func NotifyEmptyServerLogList(tip bool) {
	views.RenderInfoMessageBold("No server log files found")
	if tip {
		views.RenderTip("Use 'daytona serve' in order to create server log files")
	}
}

func NotifyEmptyRunnerList(tip bool) {
	views.RenderInfoMessageBold("No runners found")
	if tip {
		views.RenderTip("Use 'daytona runner create' to register a runner")
	}
}
