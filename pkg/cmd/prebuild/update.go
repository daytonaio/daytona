// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/add"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var prebuildUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a prebuild configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var prebuildAddView add.PrebuildAddView
		var prebuild *apiclient.PrebuildDTO
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		userGitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		if len(userGitProviders) == 0 {
			views.RenderInfoMessage("No registered Git providers have been found - please register a Git provider using 'daytona git-provider add' in order to start using prebuilds.")
			return
		}

		if len(args) < 2 {
			var prebuilds []apiclient.PrebuildDTO
			var res *http.Response
			var selectedProjectConfigName string

			if len(args) == 1 {
				selectedProjectConfigName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForProjectConfig(context.Background(), selectedProjectConfigName).Execute()
				if err != nil {
					log.Fatal(apiclient_util.HandleErrorResponse(res, err))
				}
			} else {
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Execute()
				if err != nil {
					log.Fatal(apiclient_util.HandleErrorResponse(res, err))
				}
			}

			if len(prebuilds) == 0 {
				views.RenderInfoMessage("No prebuilds found")
				return
			}

			prebuild = selection.GetPrebuildFromPrompt(prebuilds, "Update")
			if prebuild == nil {
				return
			}
		} else {
			prebuild, res, err = apiClient.PrebuildAPI.GetPrebuild(ctx, args[0], args[1]).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}
		}

		prebuildAddView.Branch = prebuild.Branch
		prebuildAddView.Retention = strconv.Itoa(int(prebuild.Retention))
		prebuildAddView.ProjectConfigName = prebuild.ProjectConfigName

		if prebuild.CommitInterval != nil {
			prebuildAddView.CommitInterval = strconv.Itoa(int(*prebuild.CommitInterval))
		}
		if len(prebuild.TriggerFiles) > 0 {
			prebuildAddView.TriggerFiles = prebuild.TriggerFiles
		}

		add.PrebuildCreationView(&prebuildAddView, true)

		var commitInterval int
		if prebuildAddView.CommitInterval != "" {
			commitInterval, err = strconv.Atoi(prebuildAddView.CommitInterval)
			if err != nil {
				log.Fatal("commit interval must be a number")
			}
		}

		retention, err := strconv.Atoi(prebuildAddView.Retention)
		if err != nil {
			log.Fatal("retention must be a number")
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
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Prebuild updated successfully")

		if runFlag {
			buildId, res, err := apiClient.BuildAPI.CreateBuild(ctx).CreateBuildDto(apiclient.CreateBuildDTO{
				ProjectConfigName: prebuildAddView.ProjectConfigName,
				Branch:            &prebuildAddView.Branch,
				PrebuildId:        &prebuildId,
			}).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			views.RenderViewBuildLogsMessage(buildId)
		}
	},
}

func init() {
	prebuildUpdateCmd.Flags().BoolVar(&runFlag, "run", false, "Run the prebuild once after updating it")
}
