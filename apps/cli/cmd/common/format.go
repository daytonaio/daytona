// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

const (
	formatFlagDescription = `Output format. Must be one of (tsv, yaml, json). Defaults to tsv when stdout is piped.`
	formatFlagName        = "format"
	formatFlagShortHand   = "f"
)

var standardOut *os.File

func init() {
	cobra.OnInitialize(resolveOutputMode)
}

// resolveOutputMode adapts CLI output to the terminal context. When stdout is
// not a TTY, FormatFlag defaults to "tsv" so piped consumers (grep, awk, ...)
// get machine-friendly rows. NO_COLOR and non-TTY stdout both strip ANSI from
// styled output.
func resolveOutputMode() {
	applyDefaults(
		term.IsTerminal(int(os.Stdout.Fd())),
		os.Getenv("NO_COLOR"),
	)
}

// applyDefaults is the pure-function core of resolveOutputMode; isolated so
// tests can exercise the precedence rules without touching real fds or env.
func applyDefaults(isStdoutTTY bool, noColor string) {
	if internal.FormatFlag == "" && !isStdoutTTY {
		internal.FormatFlag = "tsv"
	}
	if !isStdoutTTY || noColor != "" {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

type outputFormatter struct {
	data      interface{}
	formatter Formatter
}

func NewFormatter(data interface{}) *outputFormatter {
	var formatter Formatter
	switch internal.FormatFlag {
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

	formattedOutput, err := f.formatter.Format(f.data)
	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
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
	cmd.Flags().StringVarP(&internal.FormatFlag, formatFlagName, formatFlagShortHand, internal.FormatFlag, formatFlagDescription)
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		if internal.IsStructuredOutput() {
			BlockStdOut()
			// When a structured output format is requested, suppress
			// noisy warnings such as version mismatch so scripts
			// consuming json/yaml aren't broken.
			internal.SuppressVersionMismatchWarning = true
		} else {
			internal.SuppressVersionMismatchWarning = false
		}
	}
}
