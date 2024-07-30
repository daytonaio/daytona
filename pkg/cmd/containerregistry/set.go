// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"context"
	"net/url"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	containerregistry_view "github.com/daytonaio/daytona/pkg/views/containerregistry"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var containerRegistrySetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set container registry",
	Args:    cobra.NoArgs,
	Aliases: []string{"add", "update", "register"},
	Run: func(cmd *cobra.Command, args []string) {
		var registryDto *apiclient.ContainerRegistry
		selectedServer := serverFlag

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		registryView := containerregistry_view.RegistryView{
			Server:   serverFlag,
			Username: usernameFlag,
			Password: passwordFlag,
		}

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		containerRegistries, res, err := apiClient.ContainerRegistryAPI.ListContainerRegistries(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		if serverFlag == "" || usernameFlag == "" || passwordFlag == "" {
			if len(containerRegistries) == 0 {
				containerregistry_view.RegistryCreationView(&registryView, containerRegistries, false)
				selectedServer = registryView.Server
			} else {
				registryDto, err := containerregistry_view.GetRegistryFromPrompt(containerRegistries, activeProfile.Name, true)
				if err != nil {
					if common.IsCtrlCAbort(err) {
						return
					} else {
						log.Fatal(err)
					}
				}

				editing := true
				selectedServer = *registryDto.Server

				if *registryDto.Server == containerregistry_view.NewRegistryServerIdentifier {
					editing = false
					registryView.Server, registryView.Username, registryView.Password = "", "", ""
				} else {
					registryView.Server = *registryDto.Server
					registryView.Username = *registryDto.Username
					registryView.Password = *registryDto.Password
				}

				containerregistry_view.RegistryCreationView(&registryView, containerRegistries, editing)
			}
		}

		registryDto = &apiclient.ContainerRegistry{
			Server:   &registryView.Server,
			Username: &registryView.Username,
			Password: &registryView.Password,
		}

		res, err = apiClient.ContainerRegistryAPI.SetContainerRegistry(context.Background(), url.QueryEscape(selectedServer)).ContainerRegistry(*registryDto).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Registry set successfully")
	},
}

var serverFlag string
var usernameFlag string
var passwordFlag string

func init() {
	containerRegistrySetCmd.Flags().StringVarP(&serverFlag, "server", "s", "", "Server")
	containerRegistrySetCmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username")
	containerRegistrySetCmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password")
}
