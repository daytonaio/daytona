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
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/spf13/cobra"
)

var prebuildUpdateCmd = &cobra.Command{
	Use:   "update [PROJECT_CONFIG] [PREBUILD_ID]",
	Short: "Update a prebuild configuration",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuildAddView add.PrebuildAddView
		var prebuild *apiclient.PrebuildDTO
		var projectConfigRecieved string
		var retention int
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

			projectConfigRecieved = args[0]
			prebuildID := args[1]

			prebuild, res, err = apiClient.PrebuildAPI.GetPrebuild(ctx, projectConfigRecieved, prebuildID).Execute()
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
			prebuildAddView.Branch = prebuild.Branch
			prebuildAddView.Retention = strconv.Itoa(int(prebuild.Retention))
			prebuildAddView.ProjectConfigName = projectConfigRecieved
			prebuildAddView.TriggerFiles = prebuild.TriggerFiles
			prebuildAddView.CommitInterval = strconv.Itoa(int(*prebuild.CommitInterval))
			retention = int(prebuild.Retention)
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

			prebuild = selection.GetPrebuildFromPrompt(prebuilds, "Update")
			if prebuild == nil {
				return nil
			}

			projectConfigRecieved = prebuild.ProjectConfigName
			prebuildAddView = add.PrebuildAddView{
				Branch:            prebuild.Branch,
				Retention:         strconv.Itoa(int(prebuild.Retention)),
				ProjectConfigName: projectConfigRecieved,
			}
			retention, err = strconv.Atoi(prebuildAddView.Retention)
			if err != nil {
				return errors.New("retention must be a number")
			}

			if prebuild.CommitInterval != nil {
				prebuildAddView.CommitInterval = strconv.Itoa(int(*prebuild.CommitInterval))
			}
			if len(prebuild.TriggerFiles) > 0 {
				prebuildAddView.TriggerFiles = prebuild.TriggerFiles
			}
			add.PrebuildCreationView(&prebuildAddView, false)
		}

		prebuildAddView.RunBuildOnAdd = runFlag

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

var (
	branchFlag         string
	retentionFlag      int
	commitIntervalFlag int
	triggerFilesFlag   []string
	runFlag            bool
)

func init() {
	prebuildUpdateCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "Git branch for the prebuild")
	prebuildUpdateCmd.Flags().IntVarP(&retentionFlag, "retention", "r", 0, "Maximum number of resulting builds stored at a time")
	prebuildUpdateCmd.Flags().IntVarP(&commitIntervalFlag, "commit-interval", "c", 0, "Commit interval for running a prebuild - leave blank to ignore push events")
	prebuildUpdateCmd.Flags().StringSliceVarP(&triggerFilesFlag, "trigger-files", "t", nil, "Full paths of files whose changes should explicitly trigger a  prebuild")
	prebuildUpdateCmd.Flags().BoolVar(&runFlag, "run", false, "Run the prebuild once after updating it")
}
