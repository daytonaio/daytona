// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func NewTargetNameInput(targetName *string, existingTargetNames []string) error {
	input := huh.NewInput().
		Title("Name").
		Value(targetName).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("Name cannot be empty")
			}
			if slices.Contains(existingTargetNames, s) {
				return errors.New("Target with the same name already exists")
			}
			return nil
		})

	form := huh.NewForm(huh.NewGroup(input)).WithTheme(views.GetCustomTheme())
	err := form.Run()
	if err != nil {
		return err
	}
	return nil
}

func SetTargetForm(target *serverapiclient.ProviderTarget, targetManifest map[string]serverapiclient.ProviderProviderTargetProperty) error {
	fields := make([]huh.Field, 0, len(targetManifest))
	groups := []*huh.Group{}
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
		case serverapiclient.ProviderTargetPropertyTypeFilePath:
			group, value := getFilePicker(name, property)
			groups = append(groups, group...)
			options[name] = value
		}
	}

	form := huh.NewForm(append([]*huh.Group{huh.NewGroup(fields...)}, groups...)...).WithTheme(views.GetCustomTheme())
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
		case serverapiclient.ProviderTargetPropertyTypeFilePath:
			if *options[name].(*string) == "none" {
				delete(options, name)
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

func getFilePicker(name string, property serverapiclient.ProviderProviderTargetProperty) ([]*huh.Group, *string) {
	dirPath := "~"

	if property.DefaultValue != nil {
		dirPath = *property.DefaultValue
	}

	home := os.Getenv("HOME")
	if home != "" {
		dirPath = strings.Replace(dirPath, "~", home, 1)
	}

	options := []huh.Option[string]{}

	files, err := os.ReadDir(dirPath)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			options = append(options, huh.NewOption(file.Name(), path.Join(dirPath, file.Name())))
		}
	}

	customPathInput := huh.NewInput().
		Title(name).
		Validate(func(filePath string) error {
			fileInfo, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return errors.New("file does not exist")
			} else if err != nil {
				return err
			}

			if fileInfo.IsDir() {
				return errors.New("file is a directory")
			}

			return nil
		})

	if len(options) == 0 {
		return []*huh.Group{}, nil
	}

	options = append(options, huh.NewOption("Custom path", "custom-path"))
	options = append(options, huh.NewOption("None", "none"))

	var value *string = new(string)

	return []*huh.Group{
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(name).
				Options(options...).
				Value(value),
		),
		huh.NewGroup(customPathInput).WithHideFunc(func() bool {
			return *value != "custom-path"
		}),
	}, value
}
