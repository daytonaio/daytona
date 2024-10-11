// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"errors"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/build"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/add"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var (
	branchFlag         string
	retentionFlag      int
	commitIntervalFlag int
	triggerFilesFlag   []string
	projectConfigFlag  string
	runOnUpdateFlag    bool
)

var prebuildUpdateCmd = &cobra.Command{
	Use:   "update [PROJECT_CONFIG] [PREBUILD_ID]",
	Short: "Update a prebuild configuration",
	Args:  cobra.MaximumNArgs(2), // Allow up to 2 arguments
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuildAddView add.PrebuildAddView
		var prebuild *apiclient.PrebuildDTO
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		// Fetch the list of registered Git providers
		userGitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(userGitProviders) == 0 {
			views.RenderInfoMessage("No registered Git providers have been found - please register a Git provider using 'daytona git-provider add' in order to start using prebuilds.")
			return nil
		}

		// Determine the mode of operation: interactive or non-interactive
		if len(args) == 2 || (branchFlag != "" || retentionFlag != 0 || commitIntervalFlag != 0 || len(triggerFilesFlag) > 0) {
			// Non-interactive mode: use provided arguments and flags
			if len(args) < 2 {
				return errors.New("Both project config name and prebuild ID must be specified when using flags")
			}

			projectConfigFlag = args[0]
			prebuildID := args[1]

			prebuild, res, err = apiClient.PrebuildAPI.GetPrebuild(ctx, projectConfigFlag, prebuildID).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			// Set prebuild details based on flags
			if branchFlag != "" {
				prebuild.Branch = branchFlag
			}

			if retentionFlag > 0 {
				prebuild.Retention = int32(retentionFlag)
			}

			if commitIntervalFlag > 0 {
				prebuild.CommitInterval = util.Pointer(int32(commitIntervalFlag))
			}

			if len(triggerFilesFlag) > 0 {
				prebuild.TriggerFiles = triggerFilesFlag
			}

			prebuildAddView = add.PrebuildAddView{
				Branch:            prebuild.Branch,
				Retention:         strconv.Itoa(int(prebuild.Retention)),
				ProjectConfigName: projectConfigFlag,
				TriggerFiles:      triggerFilesFlag,
				CommitInterval:    strconv.Itoa(int(commitIntervalFlag)),
			}

		} else {
			// Interactive mode: Prompt for details
			var prebuilds []apiclient.PrebuildDTO
			var selectedProjectConfigName string

			if len(args) == 1 {
				selectedProjectConfigName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForProjectConfig(ctx, selectedProjectConfigName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			} else {
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuilds(ctx).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			}

			if len(prebuilds) == 0 {
				views.RenderInfoMessage("No prebuilds found")
				return nil
			}

			// Select prebuild from the prompt
			prebuild = selection.GetPrebuildFromPrompt(prebuilds, "Update")
			if prebuild == nil {
				return nil
			}

			projectConfigFlag = prebuild.ProjectConfigName
			prebuildAddView = add.PrebuildAddView{
				Branch:            prebuild.Branch,
				Retention:         strconv.Itoa(int(prebuild.Retention)),
				ProjectConfigName: projectConfigFlag,
			}

			if prebuild.CommitInterval != nil {
				prebuildAddView.CommitInterval = strconv.Itoa(int(*prebuild.CommitInterval))
			}
			if len(prebuild.TriggerFiles) > 0 {
				prebuildAddView.TriggerFiles = prebuild.TriggerFiles
			}
		}

		prebuildAddView.RunBuildOnAdd = runOnUpdateFlag

		// Confirm updates with the user
		add.PrebuildCreationView(&prebuildAddView, true)

		var commitInterval int
		if prebuildAddView.CommitInterval != "" {
			commitInterval, err = strconv.Atoi(prebuildAddView.CommitInterval)
			if err != nil {
				return errors.New("commit interval must be a number")
			}
		}

		newPrebuild := apiclient.CreatePrebuildDTO{
			Id:        &prebuild.Id,
			Branch:    &prebuildAddView.Branch,
			Retention: int32(retentionFlag),
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

		views.RenderInfoMessage("Prebuild updated successfully")

		if prebuildAddView.RunBuildOnAdd {
			projectConfig, res, err := apiClient.ProjectConfigAPI.GetProjectConfig(ctx, prebuildAddView.ProjectConfigName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			buildId, err := build.CreateBuild(apiClient, projectConfig, *newPrebuild.Branch, &prebuildId)
			if err != nil {
				return err
			}

			views.RenderViewBuildLogsMessage(buildId)
		}

		return nil
	},
}

func init() {
	prebuildUpdateCmd.Flags().StringVar(&branchFlag, "branch", "", "Git branch for the prebuild")
	prebuildUpdateCmd.Flags().IntVar(&retentionFlag, "retention", 0, "Retention period for the prebuild")
	prebuildUpdateCmd.Flags().IntVar(&commitIntervalFlag, "commit-interval", 0, "Commit interval for the prebuild")
	prebuildUpdateCmd.Flags().StringSliceVar(&triggerFilesFlag, "trigger-files", nil, "Files that trigger the prebuild")
	prebuildUpdateCmd.Flags().BoolVar(&runOnUpdateFlag, "run", false, "Run the prebuild once after updating it")
}
