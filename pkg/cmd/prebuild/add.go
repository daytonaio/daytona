// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"log"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views/prebuild/add"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var prebuildAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a prebuild configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var prebuildAddView add.PrebuildAddView
		var projectConfig *apiclient.ProjectConfig
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, "Prebuild")
		if projectConfig == nil {
			log.Fatal("No project config selected")
		}

		prebuildAddView.ProjectConfigName = *projectConfig.Name

		chosenBranch, err := workspace_util.GetBranchFromProjectConfig(projectConfig, apiClient, 0)
		if err != nil {
			log.Fatal(err)
		}

		prebuildAddView.Branch = chosenBranch

		add.PrebuildCreationView(&prebuildAddView, false)

		newPrebuild := apiclient.CreatePrebuildDTO{
			ProjectConfigName: &prebuildAddView.ProjectConfigName,
			Branch:            &prebuildAddView.Branch,
			CommitInterval:    util.Pointer(int32(10)),
			TriggerFiles:      []string{prebuildAddView.TriggerFiles},
			RunAtInit:         &runFlag,
		}

		res, err = apiClient.PrebuildAPI.SetPrebuild(ctx).Prebuild(newPrebuild).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}
	},
}

var runFlag bool

func init() {
	prebuildAddCmd.Flags().BoolVar(&runFlag, "run", false, "Run the prebuild once after adding it")
}
