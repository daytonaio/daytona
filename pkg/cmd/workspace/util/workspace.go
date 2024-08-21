// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/pkg/apiclient"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

type ProjectConfigurationFlags struct {
	Builder          *views_util.BuildChoice
	CustomImage      *string
	CustomImageUser  *string
	Branches         *[]string
	DevcontainerPath *string
	EnvVars          *[]string
	Manual           *bool
}

func AddProjectConfigurationFlags(cmd *cobra.Command, flags ProjectConfigurationFlags, multiProjectFlagException bool) {
	cmd.Flags().StringVar(flags.CustomImage, "custom-image", "", "Create the project with the custom image passed as the flag value; Requires setting --custom-image-user flag as well")
	cmd.Flags().StringVar(flags.CustomImageUser, "custom-image-user", "", "Create the project with the custom image user passed as the flag value; Requires setting --custom-image flag as well")
	cmd.Flags().StringVar(flags.DevcontainerPath, "devcontainer-path", "", "Automatically assign the devcontainer builder with the path passed as the flag value")
	cmd.Flags().Var(flags.Builder, "builder", fmt.Sprintf("Specify the builder (currently %s/%s/%s)", views_util.AUTOMATIC, views_util.DEVCONTAINER, views_util.NONE))
	cmd.Flags().StringArrayVar(flags.EnvVars, "env", []string{}, "Specify environment variables (e.g. --env 'KEY1=VALUE1' --env 'KEY2=VALUE2' ...')")
	cmd.Flags().BoolVar(flags.Manual, "manual", false, "Manually enter the Git repository")

	cmd.MarkFlagsMutuallyExclusive("builder", "custom-image")
	cmd.MarkFlagsMutuallyExclusive("builder", "custom-image-user")
	cmd.MarkFlagsMutuallyExclusive("devcontainer-path", "custom-image")
	cmd.MarkFlagsMutuallyExclusive("devcontainer-path", "custom-image-user")
	cmd.MarkFlagsRequiredTogether("custom-image", "custom-image-user")

	if multiProjectFlagException {
		cmd.MarkFlagsMutuallyExclusive("multi-project", "custom-image")
		cmd.MarkFlagsMutuallyExclusive("multi-project", "custom-image-user")
		cmd.MarkFlagsMutuallyExclusive("multi-project", "devcontainer-path")
		cmd.MarkFlagsMutuallyExclusive("multi-project", "builder")
		cmd.MarkFlagsMutuallyExclusive("multi-project", "env")
	}
}

func CheckAnyProjectConfigurationFlagSet(flags ProjectConfigurationFlags) bool {
	return *flags.CustomImage != "" || *flags.CustomImageUser != "" || (*flags.Branches != nil && len(*flags.Branches) > 0) || *flags.DevcontainerPath != "" || *flags.Builder != "" || len(*flags.EnvVars) > 0
}

func IsProjectRunning(workspace *apiclient.WorkspaceDTO, projectName string) bool {
	for _, project := range workspace.GetProjects() {
		if project.GetName() == projectName {
			return project.GetState().Uptime != 0
		}
	}
	return false
}

func GetProjectProviderMetadata(workspace *apiclient.WorkspaceDTO, projectName string) (string, error) {
	if workspace.Info != nil {
		for _, project := range workspace.Info.Projects {
			if project.Name == projectName {
				if project.ProviderMetadata == nil {
					return "", errors.New("project provider metadata is missing")
				}
				return *project.ProviderMetadata, nil
			}
		}
	}
	return "", nil
}
