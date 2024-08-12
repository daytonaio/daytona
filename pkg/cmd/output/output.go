// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	FormatDescription   = `Output format. Must be one of (yaml, json)`
	FormatFlagName      = "format"
	FormatFlagShortHand = "f"
)

var standardOut *os.File

type OutputFormatter struct {
	data      interface{}
	formatter Formatter
}

func NewOutputFormatter(data interface{}, format string) *OutputFormatter {
	var formatter Formatter
	switch format {
	case "json":
		formatter = JSONFormatter{}
	case "yaml":
		formatter = YAMLFormatter{}
	case "":
		formatter = nil
	default:
		formatter = JSONFormatter{} // Default to JSON
	}

	return &OutputFormatter{
		data:      data,
		formatter: formatter,
	}

}

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

func (f *OutputFormatter) Print() {

	formattedOutput, err := f.formatter.Format(f.data)
	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		os.Exit(1)
	}

	OpenStdOut()
	fmt.Println(formattedOutput)
}

func BlockStdOut() {
	if os.Stdout != nil {
		standardOut = os.Stdout
		os.Stdout = nil
	}
}

func OpenStdOut() {
	if os.Stdout == nil {
		os.Stdout = standardOut
		standardOut = nil
	}
}
