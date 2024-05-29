// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var defaultDirectory = "docs"

var generateDocsCmd = &cobra.Command{
	Use:   "generate-docs",
	Short: "Generate documentation for the Daytona CLI",
	Run: func(cmd *cobra.Command, args []string) {
		directory, err := cmd.Flags().GetString("directory")
		if err != nil {
			log.Fatal(err)
		}

		if directory == "" {
			directory = defaultDirectory
		}

		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		err = doc.GenMarkdownTree(cmd.Root(), directory)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Documentation generated at %s\n", directory)
	},
	Hidden: true,
}

func init() {
	generateDocsCmd.Flags().String("directory", "", "Directory to generate documentation into")
}
