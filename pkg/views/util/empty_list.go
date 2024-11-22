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
		views.RenderTip("Use 'daytona git-provider add' to add a Git provider")
	}
}

func NotifyEmptyTargetConfigList(tip bool) {
	views.RenderInfoMessageBold("No target configs found")
	if tip {
		views.RenderTip("Use 'daytona target-config set' to add a target config")
	}
}

func NotifyEmptyWorkspaceConfigList(tip bool) {
	views.RenderInfoMessageBold("No workspace configs found")
	if tip {
		views.RenderTip("Use 'daytona workspace-config add' to add a workspace config")
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

func NotifyEmptyContainerRegistryList(tip bool) {
	views.RenderInfoMessageBold("No container registries found")
	if tip {
		views.RenderTip("Use 'daytona container-registry add' to add a container registry")
	}
}

func NotifyEmptyProfileList(tip bool) {
	views.RenderInfoMessageBold("No profiles found")
	if tip {
		views.RenderTip("Use 'daytona profile add' to add a profile")
	}
}

func NotifyEmptyPrebuildList(tip bool) {
	views.RenderInfoMessageBold("No prebuilds found")
	if tip {
		views.RenderTip("Use 'daytona prebuild add' to add a prebuild")
	}
}

func NotifyEmptyApiKeyList(tip bool) {
	views.RenderInfoMessageBold("No API keys found")
	if tip {
		views.RenderTip("Use 'daytona api-key new' to create an API key")
	}
}

func NotifyEmptyBuildList(tip bool) {
	views.RenderInfoMessageBold("No builds found")
	if tip {
		views.RenderTip("Use 'daytona build run' to run a build or 'daytona prebuild add' to configure a prebuild rule")
	}
}

func NotifyEmptyEnvVarList(tip bool) {
	views.RenderInfoMessageBold("No server environment variables found")
	if tip {
		views.RenderTip("Use 'daytona env set' to add new server environment variables")
	}
}
