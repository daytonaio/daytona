// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var formatFlag string
var providerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed providers",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		providerList, err := apiclient.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		if formatFlag != "" {
			display := output.NewOutputFormatter(providerList, formatFlag)
			display.Print()
			return
		}

		provider.List(providerList)
	},
}

func init() {
	providerListCmd.PersistentFlags().StringVarP(&formatFlag, output.FormatFlagName, output.FormatFlagShortHand, formatFlag, output.FormatDescription)
	providerListCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if formatFlag != "" {
			output.BlockStdOut()
		}
	}
}
