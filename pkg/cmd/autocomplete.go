// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use:     "completion [bash|zsh|fish|powershell]",
	Short:   "Adds completion script for your shell enviornment",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error finding user home directory: %s\n", err)
			return
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
			fmt.Println("Unsupported shell type. Please use bash, zsh, fish, or powershell.")
			return
		}

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating completion script file: %s\n", err)
			return
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
			fmt.Printf("Error generating completion script: %s\n", err)
			return
		}

		sourceCommand := fmt.Sprintf("source %s\n", filePath)
		if shell == "powershell" {
			sourceCommand = fmt.Sprintf(". %s\n", filePath)
		}

		alreadyPresent := false
		profile, err := os.Open(profilePath)
		if err == nil {
			scanner := bufio.NewScanner(profile)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())			
				if strings.Contains(line, strings.TrimSpace(sourceCommand)) {
					alreadyPresent = true
					break
				}
			}
			profile.Close()
		}

		if !alreadyPresent {
			// Append the source command to the shell's profile file if not present
			profile, err := os.OpenFile(profilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Printf("Error opening profile file (%s): %s\n", profilePath, err)
				return
			}
			defer profile.Close()

			if _, err := profile.WriteString(sourceCommand); err != nil {
				fmt.Printf("Error writing to profile file (%s): %s\n", profilePath, err)
				return
			}
		}

		fmt.Println("Autocomplete script generated and injected successfully.")
		fmt.Printf("Please source your %s profile to apply the changes or restart your terminal.\n", shell)
		fmt.Printf("For manual sourcing, use: source %s\n", profilePath)
	},
}