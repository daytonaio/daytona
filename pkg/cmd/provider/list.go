// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

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

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(providerList)
			formattedData.Print()
			return
		}

		provider.List(providerList)
	},
}

func init() {
	format.RegisterFormatFlag(providerListCmd)
}
