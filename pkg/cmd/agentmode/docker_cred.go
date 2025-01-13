// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/goccy/go-json"
	"github.com/spf13/cobra"
)

type Credentials struct {
	Username string `json:"Username"`
	Secret   string `json:"Secret"`
}

var dockerCredCmd = &cobra.Command{
	Use:    "docker-cred get",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		cr, _, err := apiClient.ContainerRegistryAPI.FindContainerRegistry(ctx, string(input)).WorkspaceId(workspaceId).Execute()
		if err != nil {
			os.Exit(1)
		}

		creds := Credentials{
			Username: cr.Username,
			Secret:   cr.Password,
		}

		data, err := json.MarshalIndent(creds, "", " ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))

		return nil
	},
}
