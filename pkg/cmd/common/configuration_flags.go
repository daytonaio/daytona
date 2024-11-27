// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

type WorkspaceConfigurationFlags struct {
	Builder           *views_util.BuildChoice
	CustomImage       *string
	CustomImageUser   *string
	Branches          *[]string
	DevcontainerPath  *string
	EnvVars           *[]string
	Manual            *bool
	GitProviderConfig *string
}

func AddWorkspaceConfigurationFlags(cmd *cobra.Command, flags WorkspaceConfigurationFlags, multiWorkspaceFlagException bool) {
	cmd.Flags().StringVar(flags.CustomImage, "custom-image", "", "Create the workspace with the custom image passed as the flag value; Requires setting --custom-image-user flag as well")
	cmd.Flags().StringVar(flags.CustomImageUser, "custom-image-user", "", "Create the workspace with the custom image user passed as the flag value; Requires setting --custom-image flag as well")
	cmd.Flags().StringVar(flags.DevcontainerPath, "devcontainer-path", "", "Automatically assign the devcontainer builder with the path passed as the flag value")
	cmd.Flags().Var(flags.Builder, "builder", fmt.Sprintf("Specify the builder (currently %s/%s/%s)", views_util.AUTOMATIC, views_util.DEVCONTAINER, views_util.NONE))
	cmd.Flags().StringArrayVar(flags.EnvVars, "env", []string{}, "Specify environment variables (e.g. --env 'KEY1=VALUE1' --env 'KEY2=VALUE2' ...')")
	cmd.Flags().BoolVar(flags.Manual, "manual", false, "Manually enter the Git repository")
	cmd.Flags().StringVar(flags.GitProviderConfig, "git-provider-config", "", "Specify the Git provider configuration ID or alias")

	cmd.MarkFlagsMutuallyExclusive("builder", "custom-image")
	cmd.MarkFlagsMutuallyExclusive("builder", "custom-image-user")
	cmd.MarkFlagsMutuallyExclusive("devcontainer-path", "custom-image")
	cmd.MarkFlagsMutuallyExclusive("devcontainer-path", "custom-image-user")
	cmd.MarkFlagsRequiredTogether("custom-image", "custom-image-user")

	if multiWorkspaceFlagException {
		cmd.MarkFlagsMutuallyExclusive("multi-workspace", "custom-image")
		cmd.MarkFlagsMutuallyExclusive("multi-workspace", "custom-image-user")
		cmd.MarkFlagsMutuallyExclusive("multi-workspace", "devcontainer-path")
		cmd.MarkFlagsMutuallyExclusive("multi-workspace", "builder")
		cmd.MarkFlagsMutuallyExclusive("multi-workspace", "env")
	}
}

func CheckAnyWorkspaceConfigurationFlagSet(flags WorkspaceConfigurationFlags) bool {
	return *flags.GitProviderConfig != "" || *flags.CustomImage != "" || *flags.CustomImageUser != "" || *flags.DevcontainerPath != "" || *flags.Builder != "" || len(*flags.EnvVars) > 0
}
