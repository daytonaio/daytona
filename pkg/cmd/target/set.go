// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	internal_util "github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/provider"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/views"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/target"
	target_view "github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var pipeFile string

var TargetSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set provider target",
	Args:    cobra.NoArgs,
	Aliases: []string{"s", "add", "update", "register", "edit"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var isNewProvider bool
		var input []byte
		var err error

		if pipeFile == "-" {
			input, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			return handleTargetJSON(input)
		} else if pipeFile != "" {
			input, err = os.ReadFile(pipeFile)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", pipeFile, err)
			}
			return handleTargetJSON(input)
		}

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		providersManifest, err := manager.NewProviderManager(manager.ProviderManagerConfig{
			RegistryUrl: serverConfig.RegistryUrl,
		}).GetProvidersManifest()
		if err != nil {
			log.Error(err)
		}

		var latestProviders []apiclient.Provider
		if providersManifest != nil {
			providersManifestLatest := providersManifest.GetLatestVersions()
			if providersManifestLatest == nil {
				return errors.New("could not get latest provider versions")
			}

			latestProviders = provider.GetProviderListFromManifest(providersManifestLatest)
		} else {
			fmt.Println("Could not get provider manifest. Can't check for new providers to install")
		}

		providerViewList, err := provider.GetProviderViewOptions(apiClient, latestProviders, ctx)
		if err != nil {
			return err
		}

		selectedProvider, err := provider_view.GetProviderFromPrompt(providerViewList, "Choose a Provider", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			} else {
				return err
			}
		}

		if selectedProvider == nil {
			return nil
		}

		if selectedProvider.Installed != nil && !*selectedProvider.Installed {
			if providersManifest == nil {
				return errors.New("could not get providers manifest")
			}
			err = provider.InstallProvider(apiClient, *selectedProvider, providersManifest)
			if err != nil {
				return err
			}
			isNewProvider = true
		}

		targets, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		filteredTargets := []apiclient.ProviderTarget{}
		for _, t := range targets {
			if t.ProviderInfo.Name == selectedProvider.Name {
				filteredTargets = append(filteredTargets, t)
			}
		}

		var selectedTarget *target_view.TargetView

		if !isNewProvider || len(filteredTargets) > 0 {
			selectedTarget, err = target.GetTargetFromPrompt(filteredTargets, activeProfile.Name, nil, true, "Set")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}
		} else {
			selectedTarget = &target_view.TargetView{
				Name:    target.NewTargetName,
				Options: "{}",
			}
		}

		if selectedTarget.Name == target.NewTargetName {
			selectedTarget.Name = ""
			err = target.NewTargetNameInput(&selectedTarget.Name, internal_util.ArrayMap(targets, func(t apiclient.ProviderTarget) string {
				return t.Name
			}))
			if err != nil {
				return err
			}
		}

		targetManifest, res, err := apiClient.ProviderAPI.GetTargetManifest(context.Background(), selectedProvider.Name).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		err = target.SetTargetForm(selectedTarget, *targetManifest)
		if err != nil {
			return err
		}

		targetData := apiclient.CreateProviderTargetDTO{
			Name:    selectedTarget.Name,
			Options: selectedTarget.Options,
			ProviderInfo: apiclient.ProviderProviderInfo{
				Name:    selectedProvider.Name,
				Version: selectedProvider.Version,
			},
		}

		res, err = apiClient.TargetAPI.SetTarget(context.Background()).Target(targetData).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Target set successfully and will be used by default")
		return nil
	},
}

func handleTargetJSON(data []byte) error {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}
	var selectedTarget *target_view.TargetView
	err = parseJSON(data, &selectedTarget)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if selectedTarget.Name == "" {
		return errors.New("invalid input: 'name' field is required")
	}
	if selectedTarget.Options == "" {
		return errors.New("option fields are required to setup your target")
	}
	targetManifest, res, err := apiClient.ProviderAPI.GetTargetManifest(ctx, selectedTarget.ProviderInfo.Name).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}
	err = validateProperty(*targetManifest, selectedTarget)
	if err != nil {
		return err
	}
	targetData := apiclient.CreateProviderTargetDTO{
		Name:    selectedTarget.Name,
		Options: selectedTarget.Options,
		ProviderInfo: apiclient.ProviderProviderInfo{
			Name:    selectedTarget.ProviderInfo.Name,
			Version: selectedTarget.ProviderInfo.Version,
		},
	}
	res, err = apiClient.TargetAPI.SetTarget(ctx).Target(targetData).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}
	views.RenderInfoMessage("Target set successfully and will be used by default")
	return nil
}

func parseJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}
	return errors.New("input is not a valid JSON")
}

func validateProperty(targetManifest map[string]apiclient.ProviderProviderTargetProperty, target *target_view.TargetView) error {
	optionMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(target.Options), &optionMap); err != nil {
		return fmt.Errorf("failed to parse options JSON: %w", err)
	}

	sortedKeys := make([]string, 0, len(targetManifest))
	for k := range targetManifest {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, name := range sortedKeys {
		property := targetManifest[name]
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(target.Name)); err == nil && matched {
				continue
			}
		}
		switch *property.Type {
		case apiclient.ProviderTargetPropertyTypeFloat, apiclient.ProviderTargetPropertyTypeInt:
			_, isNumber := optionMap[name].(float64)
			if !isNumber {
				return fmt.Errorf("invalid type for %s, expected number", name)
			}

		case apiclient.ProviderTargetPropertyTypeString:
			_, isString := optionMap[name].(string)
			if !isString {
				return fmt.Errorf("invalid type for %s, expected string", name)
			}

		case apiclient.ProviderTargetPropertyTypeBoolean:
			_, isBool := optionMap[name].(bool)
			if !isBool {
				return fmt.Errorf("invalid type for %s, expected boolean", name)
			}

		case apiclient.ProviderTargetPropertyTypeOption:
			_, isString := optionMap[name].(string)
			if !isString {
				return fmt.Errorf("invalid type for %s, expected string for option", name)
			}

		case apiclient.ProviderTargetPropertyTypeFilePath:
			_, isString := optionMap[name].(string)
			if !isString {
				return fmt.Errorf("invalid type for %s, expected file path string", name)
			}

		default:
			return fmt.Errorf("unsupported provider type: %s", *property.Type)
		}
	}
	return nil
}

func init() {
	TargetSetCmd.Flags().StringVarP(&pipeFile, "file", "f", "", "Path to JSON file for target configuration, use '-' to read from stdin")
}
