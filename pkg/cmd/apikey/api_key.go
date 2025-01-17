// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var ApiKeyCmd = &cobra.Command{
	Use:     "api-key",
	Short:   "Api Key commands",
	Args:    cobra.NoArgs,
	GroupID: util.SERVER_GROUP,
	Aliases: []string{"api-keys"},
}

func init() {
	ApiKeyCmd.AddCommand(createCmd)
	ApiKeyCmd.AddCommand(deleteCmd)
	ApiKeyCmd.AddCommand(listCmd)
}
