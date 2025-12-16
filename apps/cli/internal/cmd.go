// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package internal

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	USER_GROUP    = "user"
	SANDBOX_GROUP = "sandbox"
)

// DefaultSuggestionsMinDistance is the Levenshtein distance threshold for suggestions.
// Typos within this distance will trigger "Did you mean?" suggestions.
const DefaultSuggestionsMinDistance = 2

// GetParentCmdRunE returns a RunE function for parent commands that shows suggestions
// when an unknown subcommand is provided. This enables "Did you mean?" functionality
// for typos in subcommands (e.g., "daytona sandbox lst" -> "Did you mean list?").
func GetParentCmdRunE() func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if len(args) > 0 {
			// Set suggestion distance threshold for Cobra's SuggestionsFor
			if c.SuggestionsMinimumDistance <= 0 {
				c.SuggestionsMinimumDistance = DefaultSuggestionsMinDistance
			}

			// Build error message with suggestions using Cobra's SuggestionsFor
			unknownCmd := args[0]
			errMsg := fmt.Sprintf("unknown command %q for %q", unknownCmd, c.CommandPath())

			// Get suggestions from Cobra (uses Levenshtein distance)
			suggestions := c.SuggestionsFor(unknownCmd)
			if len(suggestions) > 0 {
				errMsg += "\n\nDid you mean this?\n"
				for _, s := range suggestions {
					errMsg += fmt.Sprintf("\t%s\n", s)
				}
			}

			return fmt.Errorf("%s", errMsg)
		}
		return c.Help()
	}
}
