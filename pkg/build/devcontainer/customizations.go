// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package devcontainer

type Customizations struct {
	Extensions []string               `json:"extensions"`
	Settings   map[string]interface{} `json:"settings"`
}

type Tool string

var (
	Vscode     Tool = "vscode"
	Browser    Tool = "browser"
	Codespaces Tool = "codespaces"
)

func (c *Configuration) GetCustomizations(tool Tool) *Customizations {
	if c.Customizations == nil {
		return nil
	}

	customizations := []Customizations{}

	// Common customizations
	customizations = append(customizations, convertToCustomizations([]interface{}{c.Customizations})...)

	customizations = append(customizations, getCustomizationsByTool(Vscode, c.Customizations)...)

	if tool == Browser {
		customizations = append(customizations, getCustomizationsByTool(Browser, c.Customizations)...)
		customizations = append(customizations, getCustomizationsByTool(Codespaces, c.Customizations)...)
	}

	return MergeCustomizations(customizations)
}

func convertToCustomizations(customizations []interface{}) []Customizations {
	result := []Customizations{}

	for _, curr := range customizations {
		c := Customizations{}

		if curr.(map[string]interface{})["extensions"] != nil {
			c.Extensions = interfaceListToStringList(curr.(map[string]interface{})["extensions"].([]interface{}))
		}

		if curr.(map[string]interface{})["settings"] != nil {
			c.Settings = curr.(map[string]interface{})["settings"].(map[string]interface{})
		}

		result = append(result, c)
	}

	return result
}

func interfaceListToStringList(list []interface{}) []string {
	result := []string{}

	for _, curr := range list {
		result = append(result, curr.(string))
	}

	return result
}

func getCustomizationsByTool(tool Tool, customizationsInterface map[string]interface{}) []Customizations {
	customizations := []Customizations{}

	if toolCustomization, ok := customizationsInterface[string(tool)]; ok {
		if singleCustomization, ok := toolCustomization.(map[string]interface{}); ok {
			customizations = append(customizations, convertToCustomizations([]interface{}{singleCustomization})...)
		} else {
			customizations = append(customizations, convertToCustomizations(toolCustomization.([]interface{}))...)
		}
	}

	return customizations
}

func MergeCustomizations(customizations []Customizations) *Customizations {
	if len(customizations) == 0 {
		return nil
	}

	result := Customizations{
		Extensions: []string{},
		Settings:   map[string]interface{}{},
	}

	extensions := make(map[string]bool)
	for _, curr := range customizations {
		if curr.Extensions != nil {
			for _, extension := range curr.Extensions {
				if !extensions[extension] {
					extensions[extension] = true
					result.Extensions = append(result.Extensions, extension)
				}
			}
		}

		for key, value := range curr.Settings {
			if result.Settings[key] == nil {
				result.Settings[key] = value
			}
		}
	}

	return &result
}
