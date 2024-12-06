// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/daytonaio/daytona/pkg/apiclient"
)

func ParseJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}
	return errors.New("input is not a valid JSON")
}

func ValidateProperty(targetManifest map[string]apiclient.ProviderProviderTargetProperty, options string) error {
	optionMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(options), &optionMap); err != nil {
		return fmt.Errorf("failed to parse options JSON: %w", err)
	}

	sortedKeys := make([]string, 0, len(targetManifest))
	for k := range targetManifest {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, name := range sortedKeys {
		property := targetManifest[name]

		// Check if the property is disabled
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(name)); err == nil && matched {
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
