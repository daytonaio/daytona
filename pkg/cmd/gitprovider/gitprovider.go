// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var GitProviderCmd = &cobra.Command{
	Use:     "git-provider",
	Short:   "Manage Git provider configs",
	Args:    cobra.NoArgs,
	GroupID: util.SERVER_GROUP,
	Aliases: []string{"git-providers", "gp"},
}

func init() {
	GitProviderCmd.AddCommand(GitProviderAddCmd)
	GitProviderCmd.AddCommand(gitProviderUpdateCmd)
	GitProviderCmd.AddCommand(gitProviderDeleteCmd)
	GitProviderCmd.AddCommand(gitProviderListCmd)
}
