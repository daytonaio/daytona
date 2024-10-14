// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/build"
	"github.com/daytonaio/daytona/pkg/cmd/projectconfig"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/add"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var (
	runOnAddFlag bool
)

var prebuildAddCmd = &cobra.Command{
	Use:     "add [PROJECT_CONFIG]",
	Short:   "Add a prebuild configuration",
	Args:    cobra.MaximumNArgs(1), // Maximum one argument allowed
	Aliases: []string{"new", "create"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuildAddView add.PrebuildAddView
		var projectConfig *apiclient.ProjectConfig
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		// If no arguments and no flags are provided, run the interactive CLI
		if len(args) == 0 && branchFlag == "" && retentionFlag == 0 &&
			commitIntervalFlag == 0 && triggerFilesFlag == nil {
			// Interactive CLI logic
			gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(gitProviders) == 0 {
				views.RenderInfoMessage("No registered Git providers have been found - please register a Git provider using 'daytona git-provider add' in order to start using prebuilds.")
				return nil
			}

			projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, true, "Prebuild")
			if projectConfig == nil {
				return errors.New("No project config selected")
			}

			if projectConfig.Name == selection.NewProjectConfigIdentifier {
				projectConfig, err = projectconfig.RunProjectConfigAddFlow(apiClient, gitProviders, ctx)
				if err != nil {
					return err
				}
				if projectConfig == nil {
					return nil
				}
			}

			prebuildAddView.ProjectConfigName = projectConfig.Name
			if projectConfig.BuildConfig == nil {
				return errors.New("The chosen project config does not have a build configuration")
			}

			chosenBranch, err := workspace_util.GetBranchFromProjectConfig(projectConfig, apiClient, 0)
			if err != nil {
				return err
			}

			if chosenBranch == nil {
				fmt.Println("Operation canceled")
				return nil
			}
			prebuildAddView.RunBuildOnAdd = runOnAddFlag
			prebuildAddView.Branch = chosenBranch.Name
			add.PrebuildCreationView(&prebuildAddView, false)
		} else {
			// Non-interactive mode: use provided arguments and flags
			if len(args) > 0 {
				prebuildAddView.ProjectConfigName = args[0]

				// Fetch the project configuration based on the provided argument
				projectConfigTemp, res, err := apiClient.ProjectConfigAPI.GetProjectConfig(ctx, prebuildAddView.ProjectConfigName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
				if projectConfigTemp == nil {
					return errors.New("Invalid project config specified")
				}

				prebuildAddView.ProjectConfigName = projectConfigTemp.Name
				projectConfig = projectConfigTemp

			} else {
				return errors.New("Project config must be specified when using flags")
			}

			// Validate and handle required flags
			if branchFlag == "" {
				return errors.New("Branch flag is required when using flags")
			}
			prebuildAddView.Branch = branchFlag

			if retentionFlag <= 0 {
				return errors.New("Retention must be a positive integer")
			}
			prebuildAddView.Retention = strconv.Itoa(retentionFlag)

			if commitIntervalFlag > 0 {
				prebuildAddView.CommitInterval = strconv.Itoa(commitIntervalFlag)
			}

			prebuildAddView.TriggerFiles = triggerFilesFlag
			prebuildAddView.RunBuildOnAdd = runOnAddFlag
		}

		// Shared logic to create the prebuild configuration
		var commitInterval int
		if prebuildAddView.CommitInterval != "" {
			commitInterval, err = strconv.Atoi(prebuildAddView.CommitInterval)
			if err != nil {
				return errors.New("commit interval must be a number")
			}
		}
		var retention int

		if prebuildAddView.Retention != "" {
			retention, err = strconv.Atoi(prebuildAddView.Retention)
			if err != nil {
				return errors.New("retention must be a number")
			}
		}

		newPrebuild := apiclient.CreatePrebuildDTO{
			Branch:    &prebuildAddView.Branch,
			Retention: int32(retention),
		}

		if commitInterval != 0 {
			newPrebuild.CommitInterval = util.Pointer(int32(commitInterval))
		}

		if len(prebuildAddView.TriggerFiles) > 0 {
			newPrebuild.TriggerFiles = prebuildAddView.TriggerFiles
		}

		prebuildId, res, err := apiClient.PrebuildAPI.SetPrebuild(ctx, prebuildAddView.ProjectConfigName).Prebuild(newPrebuild).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Prebuild added successfully")

		if prebuildAddView.RunBuildOnAdd {
			buildId, err := build.CreateBuild(apiClient, projectConfig, prebuildAddView.Branch, &prebuildId)
			if err != nil {
				return err
			}

			views.RenderViewBuildLogsMessage(buildId)
		}

		return nil
	},
}

func init() {
	prebuildAddCmd.Flags().BoolVar(&runOnAddFlag, "run", false, "Run the prebuild once after adding it")
	prebuildAddCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "Git branch for the prebuild")
	prebuildAddCmd.Flags().IntVarP(&retentionFlag, "retention", "r", 0, "Maximum number of resulting builds stored at a time")
	prebuildAddCmd.Flags().IntVarP(&commitIntervalFlag, "commit-interval", "c", 0, "Commit interval for running a prebuild - leave blank to ignore push events")
	prebuildAddCmd.Flags().StringSliceVarP(&triggerFilesFlag, "trigger-files", "t", nil, "Full paths of files whose changes should explicitly trigger a  prebuild")
}
