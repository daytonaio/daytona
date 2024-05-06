// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"github.com/spf13/cobra"
)

var ApiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Api Key commands",
	Args:  cobra.NoArgs,
}

func init() {
	ApiKeyCmd.AddCommand(GenerateCmd)
	ApiKeyCmd.AddCommand(revokeCmd)
	ApiKeyCmd.AddCommand(listCmd)
}
