// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var FormatFlag string
var Output interface{}

type Formatter interface {
	Format(data interface{}) (string, error)
}

type JSONFormatter struct{}

func (f JSONFormatter) Format(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ") // Indent with two spaces
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

type YAMLFormatter struct{}

func (f YAMLFormatter) Format(data interface{}) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

func Print(data interface{}, format string) {
	var formatter Formatter

	if data == nil {
		return
	}

	switch format {
	case "json":
		formatter = JSONFormatter{}
	case "yaml":
		formatter = YAMLFormatter{}
	case "":
		return
	default:
		formatter = JSONFormatter{} // Default to JSON
	}

	formattedOutput, err := formatter.Format(data)
	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(formattedOutput)
}
