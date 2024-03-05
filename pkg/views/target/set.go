// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
)

func NewTargetNameInput(targetName *string) error {
	input := huh.NewInput().
		Title("Name").
		Value(targetName).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("Name cannot be empty")
			}
			return nil
		})

	form := huh.NewForm(huh.NewGroup(input))
	err := form.Run()
	if err != nil {
		return err
	}
	return nil
}

func SetTargetForm(target *serverapiclient.ProviderTarget, targetManifest map[string]serverapiclient.ProviderProviderTargetProperty) error {
	fields := make([]huh.Field, 0, len(targetManifest))
	options := make(map[string]interface{})

	err := json.Unmarshal([]byte(*target.Options), &options)
	if err != nil {
		return err
	}

	sortedKeys := make([]string, 0, len(targetManifest))
	for k := range targetManifest {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, name := range sortedKeys {
		property := targetManifest[name]
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(*target.Name)); err == nil && matched {
				continue
			}
		}

		switch *property.Type {
		case serverapiclient.ProviderTargetPropertyTypeFloat, serverapiclient.ProviderTargetPropertyTypeInt:
			var initialValue *string
			floatValue, ok := options[name].(float64)
			if ok {
				v := strconv.FormatFloat(floatValue, 'f', -1, 64)
				initialValue = &v
			}

			input, value := getInput(name, property, initialValue)
			fields = append(fields, input)
			options[name] = value
		case serverapiclient.ProviderTargetPropertyTypeString:
			var initialValue *string
			v, ok := options[name].(string)
			if ok {
				initialValue = &v
			}

			input, value := getInput(name, property, initialValue)
			fields = append(fields, input)
			options[name] = value
		case serverapiclient.ProviderTargetPropertyTypeBoolean:
			var initialValue *bool
			v, ok := options[name].(bool)
			if ok {
				initialValue = &v
			}

			confirm, value := getConfirm(name, property, initialValue)
			fields = append(fields, confirm)
			options[name] = value
		case serverapiclient.ProviderTargetPropertyTypeOption:
			var initialValue *string
			v, ok := options[name].(string)
			if ok {
				initialValue = &v
			}

			selectField, value := getSelect(name, property, initialValue)
			fields = append(fields, selectField)
			options[name] = value
		}
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	err = form.Run()
	if err != nil {
		return err
	}

	for name, property := range targetManifest {
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(*target.Name)); err == nil && matched {
				continue
			}
		}
		switch *property.Type {
		case serverapiclient.ProviderTargetPropertyTypeInt:
			options[name], err = strconv.Atoi(*options[name].(*string))
			if err != nil {
				return err
			}
		case serverapiclient.ProviderTargetPropertyTypeFloat:
			options[name], err = strconv.ParseFloat(*options[name].(*string), 64)
			if err != nil {
				return err
			}
		}
	}

	jsonContent, err := json.MarshalIndent(options, "", "  ")
	if err != nil {
		return err
	}
	content := string(jsonContent)

	target.Options = &content
	return nil
}

func getInput(name string, property serverapiclient.ProviderProviderTargetProperty, initialValue *string) (*huh.Input, *string) {
	value := property.DefaultValue
	if initialValue != nil {
		value = initialValue
	}

	return huh.NewInput().
		Title(name).
		Value(value).
		Password(property.InputMasked != nil && *property.InputMasked).
		Validate(func(s string) error {
			switch *property.Type {
			case serverapiclient.ProviderTargetPropertyTypeInt:
				_, err := strconv.Atoi(s)
				return err
			case serverapiclient.ProviderTargetPropertyTypeFloat:
				_, err := strconv.ParseFloat(s, 64)
				return err
			}
			return nil
		}), value
}

func getSelect(name string, property serverapiclient.ProviderProviderTargetProperty, initialValue *string) (*huh.Select[string], *string) {
	value := property.DefaultValue
	if initialValue != nil {
		value = initialValue
	}

	return huh.NewSelect[string]().
		Title(name).
		Options(util.ArrayMap(property.Options, func(o string) huh.Option[string] {
			return huh.NewOption(o, o)
		})...).
		Value(value), value
}

func getConfirm(name string, property serverapiclient.ProviderProviderTargetProperty, initialValue *bool) (*huh.Confirm, *bool) {
	value := false
	if property.DefaultValue != nil && *property.DefaultValue == "true" {
		value = true
	}
	if initialValue != nil {
		value = *initialValue
	}

	return huh.NewConfirm().
		Title(name).
		Value(&value), &value
}
