// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"github.com/spf13/cobra"
)

var GitProviderCmd = &cobra.Command{
	Use:     "git-providers",
	Aliases: []string{"git-provider", "gp"},
	Short:   "Manage Git providers",
}

func init() {
	GitProviderCmd.AddCommand(GitProviderAddCmd)
	GitProviderCmd.AddCommand(gitProviderDeleteCmd)
	GitProviderCmd.AddCommand(gitProviderListCmd)
}
