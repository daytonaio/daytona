// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/cmd/provider/target"
	"github.com/spf13/cobra"
)

var providerNameArg string

var ProviderCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage providers",
}

func init() {
	ProviderCmd.AddCommand(providerListCmd)
	ProviderCmd.AddCommand(providerUninstallCmd)
	ProviderCmd.AddCommand(providerInstallCmd)
	ProviderCmd.AddCommand(target.TargetCmd)
}
