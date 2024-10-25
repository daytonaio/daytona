// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/target/info"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show project info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.ExactArgs(0),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var target *apiclient.TargetDTO

		target, err := apiclient_util.GetTarget(targetId, true)
		if err != nil {
			return err
		}

		if target == nil {
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(target)
			formattedData.Print()
			return nil
		}

		info.Render(target, "", false)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(infoCmd)
}
