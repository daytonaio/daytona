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
	Args:    cobra.NoArgs,
	GroupID: util.SERVER_GROUP,
	Aliases: []string{"providers"},
}

func init() {
	ProviderCmd.AddCommand(listCmd)
	ProviderCmd.AddCommand(uninstallCmd)
	ProviderCmd.AddCommand(installCmd)
	ProviderCmd.AddCommand(updateCmd)
}
