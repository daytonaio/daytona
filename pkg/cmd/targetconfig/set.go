// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

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
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var pipeFile string

var TargetConfigSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set target config",
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

		targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		filteredConfigs := []apiclient.TargetConfig{}
		for _, t := range targetConfigs {
			if t.ProviderInfo.Name == selectedProvider.Name {
				filteredConfigs = append(filteredConfigs, t)
			}
		}

		var selectedTargetConfig *targetconfig.TargetConfigView

		if !isNewProvider || len(filteredConfigs) > 0 {
			selectedTargetConfig, err = targetconfig.GetTargetConfigFromPrompt(filteredConfigs, activeProfile.Name, nil, true, "Set")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}
		} else {
			selectedTargetConfig = &targetconfig.TargetConfigView{
				Name:    targetconfig.NewTargetConfigName,
				Options: "{}",
			}
		}

		if selectedTargetConfig.Name == targetconfig.NewTargetConfigName {
			selectedTargetConfig.Name = ""
			err = targetconfig.NewTargetConfigNameInput(&selectedTargetConfig.Name, internal_util.ArrayMap(targetConfigs, func(t apiclient.TargetConfig) string {
				return t.Name
			}))
			if err != nil {
				return err
			}
		}

		targetConfigManifest, res, err := apiClient.ProviderAPI.GetTargetConfigManifest(context.Background(), selectedProvider.Name).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		err = targetconfig.SetTargetConfigForm(selectedTargetConfig, *targetConfigManifest)
		if err != nil {
			return err
		}

		targetConfigData := apiclient.CreateTargetConfigDTO{
			Name:    selectedTargetConfig.Name,
			Options: selectedTargetConfig.Options,
			ProviderInfo: apiclient.ProviderProviderInfo{
				Name:    selectedProvider.Name,
				Version: selectedProvider.Version,
			},
		}

		res, err = apiClient.TargetConfigAPI.SetTargetConfig(context.Background()).TargetConfig(targetConfigData).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Target config set successfully and will be used by default")
		return nil
	},
}

func handleTargetJSON(data []byte) error {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}
	var selectedTargetConfig *targetconfig.TargetConfigView
	err = parseJSON(data, &selectedTargetConfig)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	if selectedTargetConfig.Name == "" {
		return errors.New("invalid input: 'name' field is required")
	}
	if selectedTargetConfig.Options == "" {
		return errors.New("option fields are required to setup your target")
	}
	targetManifest, res, err := apiClient.ProviderAPI.GetTargetConfigManifest(ctx, selectedTargetConfig.ProviderInfo.Name).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}
	err = validateProperty(*targetManifest, selectedTargetConfig)
	if err != nil {
		return err
	}
	targetConfigData := apiclient.CreateTargetConfigDTO{
		Name:    selectedTargetConfig.Name,
		Options: selectedTargetConfig.Options,
		ProviderInfo: apiclient.ProviderProviderInfo{
			Name:    selectedTargetConfig.ProviderInfo.Name,
			Version: selectedTargetConfig.ProviderInfo.Version,
		},
	}
	res, err = apiClient.TargetConfigAPI.SetTargetConfig(ctx).TargetConfig(targetConfigData).Execute()
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

func validateProperty(targetManifest map[string]apiclient.TargetConfigProperty, targetConfig *targetconfig.TargetConfigView) error {
	optionMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(targetConfig.Options), &optionMap); err != nil {
		return fmt.Errorf("failed to parse options JSON: %w", err)
	}
	for optionKey := range optionMap {
		if _, exists := targetManifest[optionKey]; !exists {
			return fmt.Errorf("invalid property '%s' for target manifest '%s'", optionKey, targetConfig.Name)
		}
	}

	sortedKeys := make([]string, 0, len(targetManifest))
	for k := range targetManifest {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, name := range sortedKeys {
		if _, present := optionMap[name]; !present {
			continue
		}

		property := targetManifest[name]
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(targetConfig.Name)); err == nil && matched {
				if !contains(property.Options, optionMap[name]) {
					return fmt.Errorf("unexpected property '%s' for target manifest '%s'", name, targetConfig.Name)
				}
				continue
			}
		}

		switch *property.Type {
		case apiclient.TargetConfigPropertyTypeFloat, apiclient.TargetConfigPropertyTypeInt:
			_, isNumber := optionMap[name].(float64)
			if !isNumber {
				return fmt.Errorf("invalid type for %s, expected number", name)
			}

		case apiclient.TargetConfigPropertyTypeString:
			_, isString := optionMap[name].(string)
			if !isString {
				return fmt.Errorf("invalid type for %s, expected string", name)
			}

		case apiclient.TargetConfigPropertyTypeBoolean:
			_, isBool := optionMap[name].(bool)
			if !isBool {
				return fmt.Errorf("invalid type for %s, expected boolean", name)
			}

		case apiclient.TargetConfigPropertyTypeOption:
			optionValue, ok := optionMap[name].(string)
			if !ok {
				return fmt.Errorf("invalid value for '%s': expected a string", name)
			}
			valid := false
			for _, allowedOption := range property.Options {
				if optionValue == allowedOption {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("unexpected property '%s' for target manifest '%s' : valid properties are %v", optionValue, targetConfig.Name, property.Options)
			}

		case apiclient.TargetConfigPropertyTypeFilePath:
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

func contains(slice []string, item interface{}) bool {
	for _, val := range slice {
		if val == item {
			return true
		}
	}
	return false
}

func init() {
	TargetConfigSetCmd.Flags().StringVarP(&pipeFile, "file", "f", "", "Path to JSON file for target configuration, use '-' to read from stdin")
}
