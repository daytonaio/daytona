// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new profile",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		var profileName string
		var apiUrl string
		var apiKey string

		if nameFlag != "" {
			profileName = nameFlag
		} else {
			profileName, err = view_common.PromptForInput("", "Enter profile name", "")
			if err != nil {
				return err
			}
		}

		if apiUrlFlag != "" {
			apiUrl = apiUrlFlag
		} else {
			defaultUrl := config.GetDaytonaApiUrl()
			if internal.Version == "v0.0.0-dev" {
				defaultUrl = "http://localhost:3001/api"
			}
			apiUrl, err = view_common.PromptForInput(defaultUrl, "Enter API URL", "")
			if err != nil {
				return err
			}
		}

		// Normalize API URL
		apiUrl = strings.TrimSuffix(apiUrl, "/")
		if !strings.HasSuffix(apiUrl, "/api") {
			apiUrl = apiUrl + "/api"
		}

		// Check if profile name already exists
		for _, p := range c.Profiles {
			if p.Name == profileName {
				return fmt.Errorf("profile with name %s already exists", profileName)
			}
		}

		// Check if API URL already exists
		for _, p := range c.Profiles {
			if p.Api.Url == apiUrl {
				return fmt.Errorf("profile with API URL %s already exists", apiUrl)
			}
		}

		newProfile := config.Profile{
			Id:   uuid.New().String(),
			Name: profileName,
			Api: config.ServerApi{
				Url: apiUrl,
			},
		}

		if apiKeyFlag != "" {
			apiKey = apiKeyFlag
			newProfile.Api.Key = &apiKey
		}

		err = c.AddProfile(newProfile)
		if err != nil {
			return err
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Profile %s added and set as active", profileName))
		return nil
	},
}

var (
	nameFlag   string
	apiUrlFlag string
	apiKeyFlag string
)

func init() {
	AddCmd.Flags().StringVar(&nameFlag, "name", "", "Profile name")
	AddCmd.Flags().StringVar(&apiUrlFlag, "api_url", "", "API URL")
	AddCmd.Flags().StringVar(&apiKeyFlag, "api_key", "", "API key (optional)")
}
