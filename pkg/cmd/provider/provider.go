// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var ProviderCmd = &cobra.Command{
	Use:     "provider",
	Short:   "Manage providers",
	GroupID: util.SERVER_GROUP,
}

func init() {
	ProviderCmd.AddCommand(providerListCmd)
	ProviderCmd.AddCommand(providerUninstallCmd)
	ProviderCmd.AddCommand(providerInstallCmd)
	ProviderCmd.AddCommand(providerUpdateCmd)
}
