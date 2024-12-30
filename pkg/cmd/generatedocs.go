// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var yamlDirectory = "hack"
var defaultDirectory = "docs"

var generateDocsCmd = &cobra.Command{
	Use:   "generate-docs",
	Short: "Generate documentation for the Daytona CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, err := cmd.Flags().GetString("directory")
		if err != nil {
			return err
		}

		if directory == "" {
			directory = defaultDirectory
		}

		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Join(yamlDirectory, directory), os.ModePerm)
		if err != nil {
			return err
		}

		err = doc.GenMarkdownTree(cmd.Root(), directory)
		if err != nil {
			return err
		}

		err = doc.GenYamlTree(cmd.Root(), filepath.Join(yamlDirectory, directory))
		if err != nil {
			return err
		}

		fmt.Printf("Documentation generated at %s\n", directory)
		return nil
	},
	Hidden: true,
}

func init() {
	generateDocsCmd.Flags().String("directory", "", "Directory to generate documentation into")
}
