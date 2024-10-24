// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const helpDescriptionLabelWidth = 32
const helpDescriptionLabelWidthWithSigil = 30
const sigilThresholdWidth = 115

func GetLongDescription() string {
	var response string
	if shouldDisplayASCIIArt() {
		response = getLongDescriptionFull()
	} else {
		response = getLongDescriptionText()
	}
	return response
}

func shouldDisplayASCIIArt() bool {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return false
	}

	return width >= sigilThresholdWidth
}

func getLongDescriptionText() string {
	var response string
	response += "\n" + fmt.Sprintf("  \x1b[1m%s\x1b[0m%s\n\n", "Daytona", " - your Dev Environment Manager") +
		"  Use the following commands to get started:\n\n" +
		fmt.Sprintf("  \x1b[1m%-*s\x1b[0m%s", helpDescriptionLabelWidth, "1) daytona server", "Start the Daytona Server process locally\n") +
		fmt.Sprintf("  \x1b[1m%-*s\x1b[0m%s", helpDescriptionLabelWidth, "2) daytona git-providers add", "Register a Git provider of your choice\n") +
		fmt.Sprintf("  \x1b[1m%-*s\x1b[0m%s", helpDescriptionLabelWidth, "3) daytona target-config set", "Set a target config to spin up your Dev Environments on\n") +
		fmt.Sprintf("  \x1b[1m%-*s\x1b[0m%s", helpDescriptionLabelWidth, "4) daytona ide", "Choose the default IDE\n") +
		fmt.Sprintf("  \x1b[1m%-*s\x1b[0m%s", helpDescriptionLabelWidth, "5) daytona whoami", "Show information about the currently logged in user\n") +
		fmt.Sprintf("  \n%s\x1b[1m%s\x1b[0m\n\n", "That's it! Start coding - ", "daytona create")
	return response
}

func getLongDescriptionFull() string {
	response := ""

	response +=
		"\n" +
			fmt.Sprintf("%s\n", "               @@@          ") +
			fmt.Sprintf("%s\x1b[1m%s\x1b[0m%s\n", "            @@@@@           ", "Daytona", " - your Dev Environment Manager") +
			fmt.Sprintf("%s\n", "     @@@@  @@@@@            ") +
			fmt.Sprintf("%sUse the following commands to get started:\n", "     @@@@@@@@@@@@@@@@@@@@   ") +
			fmt.Sprintf("%s\n", "     @@@@ @@  @@@@@@@@@@@   ") +
			fmt.Sprintf("%s\x1b[1m%-*s\x1b[0m%s", " @@@@@@@@        @@@        ", helpDescriptionLabelWidthWithSigil, "1) daytona server", "Start the Daytona Server process locally\n") +
			fmt.Sprintf("%s\x1b[1m%-*s\x1b[0m%s", "  @@@@@          @@@@@      ", helpDescriptionLabelWidthWithSigil, "2) daytona git-providers add", "Register a Git provider of your choice\n") +
			fmt.Sprintf("%s\x1b[1m%-*s\x1b[0m%s", "    @@@@@       @@@@@@@@    ", helpDescriptionLabelWidthWithSigil, "3) daytona target-config set", "Set a target config to spin up your Dev Environments\n") +
			fmt.Sprintf("%s\x1b[1m%-*s\x1b[0m%s", "     @@@        @@@@@@@@@   ", helpDescriptionLabelWidthWithSigil, "4) daytona ide", "Choose the default IDE\n") +
			fmt.Sprintf("%s\x1b[1m%-*s\x1b[0m%s", "  @@@@@@@@@ @@@@@@@@  @@    ", helpDescriptionLabelWidthWithSigil, "5) daytona whoami", "Show information about the currently logged in user\n") +
			fmt.Sprintf("%s\n", "  @@@@@@@@@@@@@@@@@@        ") +
			fmt.Sprintf("%s%s\x1b[1m%s\x1b[0m", "        @@@@@@  @@@@        ", "That's it! Start coding - ", "daytona create\n") +
			fmt.Sprintf("%s\n\n", "         @@@    @@@   ")

	return response
}
