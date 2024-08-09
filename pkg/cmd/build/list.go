// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"log"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var buildListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Delete a build configuration",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		fmt.Println("Test.")
		fmt.Println(len(buildList))
		fmt.Println(buildList)

		if len(buildList) == 0 {
			views.RenderInfoMessage("No builds found.")
			return
		}

		if output.FormatFlag != "" {
			output.Output = buildList
			return
		}
	},
}
