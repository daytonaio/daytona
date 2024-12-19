// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func NewTargetConfigNameInput(name *string, existingNames []string) error {
	input := huh.NewInput().
		Title("Target Config Name").
		Value(name).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("Name cannot be empty")
			}
			if slices.Contains(existingNames, s) {
				return errors.New("Target config with the same name already exists")
			}
			return nil
		})

	form := huh.NewForm(huh.NewGroup(input)).WithTheme(views.GetCustomTheme()).WithHeight(5)
	err := form.Run()
	if err != nil {
		return err
	}
	return nil
}

func SetTargetConfigForm(targetConfig *TargetConfigView, targetConfigManifest map[string]apiclient.TargetConfigProperty) error {
	fields := make([]huh.Field, 0, len(targetConfigManifest))
	groups := []*huh.Group{}
	options := make(map[string]interface{})

	err := json.Unmarshal([]byte(targetConfig.Options), &options)
	if err != nil {
		return err
	}

	sortedKeys := make([]string, 0, len(targetConfigManifest))
	for k := range targetConfigManifest {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, name := range sortedKeys {
		property := targetConfigManifest[name]
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(targetConfig.Name)); err == nil && matched {
				continue
			}
		}

		if property.Type == nil {
			continue
		}

		switch *property.Type {
		case apiclient.TargetConfigPropertyTypeFloat, apiclient.TargetConfigPropertyTypeInt:
			var initialValue *string
			floatValue, ok := options[name].(float64)
			if ok {
				v := strconv.FormatFloat(floatValue, 'f', -1, 64)
				initialValue = &v
			}

			input, value := getInput(name, property, initialValue)
			fields = append(fields, input)
			options[name] = value
		case apiclient.TargetConfigPropertyTypeString:
			var initialValue *string
			v, ok := options[name].(string)
			if ok {
				initialValue = &v
			}

			input, value := getInput(name, property, initialValue)
			fields = append(fields, input)
			options[name] = value
		case apiclient.TargetConfigPropertyTypeBoolean:
			var initialValue *bool
			v, ok := options[name].(bool)
			if ok {
				initialValue = &v
			}

			confirm, value := getConfirm(name, property, initialValue)
			fields = append(fields, confirm)
			options[name] = value
		case apiclient.TargetConfigPropertyTypeOption:
			var initialValue *string
			v, ok := options[name].(string)
			if ok {
				initialValue = &v
			}

			selectField, value := getSelect(name, property, initialValue)
			fields = append(fields, selectField)
			options[name] = value
		case apiclient.TargetConfigPropertyTypeFilePath:
			var initialValue *string
			v, ok := options[name].(string)
			if ok {
				initialValue = &v
			}
			group, value := getFilePicker(name, property, initialValue)
			groups = append(groups, group...)
			options[name] = value
		}
	}

	form := huh.NewForm(append([]*huh.Group{huh.NewGroup(fields...)}, groups...)...).WithTheme(views.GetCustomTheme()).WithProgramOptions(tea.WithAltScreen())
	err = form.Run()
	if err != nil {
		return err
	}

	for name, property := range targetConfigManifest {
		if property.DisabledPredicate != nil && *property.DisabledPredicate != "" {
			if matched, err := regexp.Match(*property.DisabledPredicate, []byte(targetConfig.Name)); err == nil && matched {
				continue
			}
		}

		if property.Type == nil {
			continue
		}

		switch *property.Type {
		case apiclient.TargetConfigPropertyTypeInt:
			options[name], err = strconv.Atoi(*options[name].(*string))
			if err != nil {
				return err
			}
		case apiclient.TargetConfigPropertyTypeFloat:
			options[name], err = strconv.ParseFloat(*options[name].(*string), 64)
			if err != nil {
				return err
			}
		case apiclient.TargetConfigPropertyTypeFilePath:
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

	targetConfig.Options = content
	return nil
}

func getInput(name string, property apiclient.TargetConfigProperty, initialValue *string) (*huh.Input, *string) {
	value := property.DefaultValue
	if initialValue != nil {
		value = initialValue
	}

	input := huh.NewInput().
		Title(name).
		Description(*property.Description).
		Value(value).
		Validate(func(s string) error {
			if property.Type == nil {
				return errors.New("property type is not defined")
			}

			switch *property.Type {
			case apiclient.TargetConfigPropertyTypeInt:
				_, err := strconv.Atoi(s)
				return err
			case apiclient.TargetConfigPropertyTypeFloat:
				_, err := strconv.ParseFloat(s, 64)
				return err
			}
			return nil
		})

	if property.InputMasked != nil && *property.InputMasked {
		input = input.EchoMode(huh.EchoModePassword)
	}

	if len(property.Suggestions) > 0 {
		input = input.Suggestions(property.Suggestions)
	}

	return input, value
}

func getSelect(name string, property apiclient.TargetConfigProperty, initialValue *string) (*huh.Select[string], *string) {
	value := property.DefaultValue
	if initialValue != nil {
		value = initialValue
	}

	return huh.NewSelect[string]().
		Title(name).
		Description(*property.Description).
		Options(util.ArrayMap(property.Options, func(o string) huh.Option[string] {
			return huh.NewOption(o, o)
		})...).
		Value(value), value
}

func getConfirm(name string, property apiclient.TargetConfigProperty, initialValue *bool) (*huh.Confirm, *bool) {
	value := false
	if property.DefaultValue != nil && *property.DefaultValue == "true" {
		value = true
	}
	if initialValue != nil {
		value = *initialValue
	}

	return huh.NewConfirm().
		Title(name).
		Description(*property.Description).
		Value(&value), &value
}

func getFilePicker(name string, property apiclient.TargetConfigProperty, initialValue *string) ([]*huh.Group, *string) {
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

			options = append(options, huh.NewOption(file.Name(), filepath.Join(dirPath, file.Name())))
		}
	}

	var value *string = new(string)
	if initialValue != nil {
		*value = *initialValue
	}

	customPathInput := huh.NewInput().
		Title(name).
		Description(*property.Description).
		Value(value).
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
		return []*huh.Group{huh.NewGroup(customPathInput)}, value
	}

	options = append(options, huh.NewOption("Custom path", "custom-path"))
	options = append(options, huh.NewOption("None", "none"))

	description := fmt.Sprintf("%s\nShowing files in: %s\nYou can select a file, choose None or enter a Custom path", *property.Description, dirPath)

	return []*huh.Group{
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(name).
				Description(description).
				Options(options...).
				Value(value).Validate(func(s string) error {
				if s == "custom-path" {
					*value = ""
				}
				return nil
			}).
				WithHeight(10 + strings.Count(description, "\n")),
		),
		huh.NewGroup(customPathInput).WithHideFunc(func() bool {
			return *value != ""
		}),
	}, value
}
