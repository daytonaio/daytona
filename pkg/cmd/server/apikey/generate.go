// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
	"github.com/daytonaio/daytona/pkg/views/util"
)

var saveFlag bool

var generateCmd = &cobra.Command{
	Use:     "generate [NAME]",
	Short:   "Generate a new API key",
	Aliases: []string{"g", "new"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var keyName string

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiKeyList, _, err := apiClient.ApiKeyAPI.ListClientApiKeys(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(nil, err))
		}

		if len(args) == 1 {
			keyName = args[0]
		} else {
			apikey.ApiKeyCreationView(&keyName, &saveFlag, apiKeyList)
		}

		for _, key := range apiKeyList {
			if *key.Name == keyName {
				log.Fatal("key name already exists, please choose a different one")
			}
		}

		key, _, err := apiClient.ApiKeyAPI.GenerateApiKey(ctx, keyName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(nil, err))
		}

		if saveFlag {
			err := saveKeyToDefaultProfile(key)
			if err != nil {
				log.Fatal(err)
			}
			util.RenderBorderedMessage("API key saved to your default profile")
			return
		}

		util.RenderBorderedMessage(fmt.Sprintf("Generated API key: %s\n\nYou can add it to a profile by running:\n\ndaytona profile edit -k %s\n\nMake sure to copy it as you will not be able to see it again.", key, key))
	},
}

func saveKeyToDefaultProfile(key string) error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	for _, p := range c.Profiles {
		if p.Id == "default" {
			p.Api.Key = key
			return c.EditProfile(p)
		}
	}

	return fmt.Errorf("default profile not found")
}

func init() {
	generateCmd.Flags().BoolVarP(&saveFlag, "save", "s", false, "Save the API key to your default profile on this machine")
}
