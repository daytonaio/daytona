// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var AutoCompleteCmd = &cobra.Command{
	Use:   "autocomplete [bash|zsh|fish|powershell]",
	Short: "Adds completion script for your shell enviornment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error finding user home directory: %s", err)
		}

		var filePath, profilePath string
		switch shell {
		case "bash":
			filePath = filepath.Join(homeDir, ".daytona.completion_script.bash")
			profilePath = filepath.Join(homeDir, ".bashrc")
		case "zsh":
			filePath = filepath.Join(homeDir, ".daytona.completion_script.zsh")
			profilePath = filepath.Join(homeDir, ".zshrc")
		case "fish":
			filePath = filepath.Join(homeDir, ".config", "fish", "daytona.completion_script.fish")
			profilePath = filepath.Join(homeDir, ".config", "fish", "config.fish")
		case "powershell":
			filePath = filepath.Join(homeDir, "daytona.completion_script.ps1")
			profilePath = filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
		default:
			return errors.New("unsupported shell type. Please use bash, zsh, fish, or powershell")
		}

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating completion script file: %s", err)
		}
		defer file.Close()

		switch shell {
		case "bash":
			err = cmd.Root().GenBashCompletion(file)
		case "zsh":
			err = cmd.Root().GenZshCompletion(file)
		case "fish":
			err = cmd.Root().GenFishCompletion(file, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(file)
		}

		if err != nil {
			return fmt.Errorf("error generating completion script: %s", err)
		}

		sourceCommand := fmt.Sprintf("\nsource %s\n", filePath)
		if shell == "powershell" {
			sourceCommand = fmt.Sprintf(". %s\n", filePath)
		}

		alreadyPresent := false
		// Read existing content from the file
		profile, err := os.ReadFile(profilePath)

		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("error while reading profile (%s): %s\n", profilePath, err)
		}

		if strings.Contains(string(profile), strings.TrimSpace(sourceCommand)) {
			alreadyPresent = true
		}

		if !alreadyPresent {
			// Append the source command to the shell's profile file if not present
			profile, err := os.OpenFile(profilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				return fmt.Errorf("error opening profile file (%s): %s", profilePath, err)
			}
			defer profile.Close()

			if _, err := profile.WriteString(sourceCommand); err != nil {
				return fmt.Errorf("error writing to profile file (%s): %s", profilePath, err)
			}
		}

		fmt.Println("Autocomplete script generated and injected successfully.")
		fmt.Printf("Please source your %s profile to apply the changes or restart your terminal.\n", shell)
		fmt.Printf("For manual sourcing, use: source %s\n", profilePath)
		if shell == "bash" {
			fmt.Println("Please make sure that you have bash-completion installed in order to get full autocompletion functionality.")
			fmt.Println("On how to install bash-completion, please refer to the following link: https://www.daytona.io/docs/tools/cli/#daytona-autocomplete")
		}

		return nil
	},
}
