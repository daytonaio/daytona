// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	formatFlagDescription = `Output format. Must be one of (yaml, json)`
	formatFlagName        = "format"
	formatFlagShortHand   = "f"
)

var (
	FormatFlag  string
	standardOut *os.File
)

type outputFormatter struct {
	data      interface{}
	formatter Formatter
}

func NewFormatter(data interface{}) *outputFormatter {
	var formatter Formatter
	switch FormatFlag {
	case "json":
		formatter = JSONFormatter{}
	case "yaml":
		formatter = YAMLFormatter{}
	case "":
		formatter = nil
	default:
		formatter = JSONFormatter{} // Default to JSON
	}

	return &outputFormatter{
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

func (f *outputFormatter) Print() {
	// No --format means no formatter; printing structured output is a no-op.
	if f.formatter == nil {
		return
	}

	formattedOutput, err := f.formatter.Format(f.data)
	if err != nil {
		// Stdout is blocked while a structured format is active, so report
		// the formatting failure on stderr.
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	UnblockStdOut()
	fmt.Println(formattedOutput)
	BlockStdOut()
}

func BlockStdOut() {
	if os.Stdout != nil {
		standardOut = os.Stdout
		os.Stdout = nil
	}
}

func UnblockStdOut() {
	if os.Stdout == nil {
		os.Stdout = standardOut
		standardOut = nil
	}
}

func RegisterFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&FormatFlag, formatFlagName, formatFlagShortHand, FormatFlag, formatFlagDescription)
	chainFormatPreRunE(cmd)
}

// RegisterFormatFlagNoShorthand registers --format without the -f shorthand,
// for commands where -f is already taken (e.g. --force, --dockerfile).
func RegisterFormatFlagNoShorthand(cmd *cobra.Command) {
	cmd.Flags().StringVar(&FormatFlag, formatFlagName, FormatFlag, formatFlagDescription)
	chainFormatPreRunE(cmd)
}

// chainFormatPreRunE installs the --format validation as the command's
// PreRunE, calling any pre-existing PreRunE after validation succeeds instead
// of clobbering it.
func chainFormatPreRunE(cmd *cobra.Command) {
	original := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := formatFlagPreRunE(cmd, args); err != nil {
			return err
		}
		if original != nil {
			return original(cmd, args)
		}
		return nil
	}
}

func formatFlagPreRunE(cmd *cobra.Command, args []string) error {
	switch FormatFlag {
	case "":
		internal.SuppressVersionMismatchWarning = false
	case "json", "yaml":
		BlockStdOut()
		// When a structured output format is requested, suppress
		// noisy warnings such as version mismatch so scripts
		// consuming json/yaml aren't broken.
		internal.SuppressVersionMismatchWarning = true
	default:
		return clierr.Newf(clierr.CategoryUsage, "invalid --format value %q: must be one of json, yaml", FormatFlag)
	}
	return nil
}
