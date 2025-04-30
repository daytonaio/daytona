// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/internal"
	"github.com/spf13/cobra"
)

var SandboxCmd = &cobra.Command{
	Use:     "sandbox",
	Short:   "Manage Daytona sandboxes",
	Long:    "Commands for managing Daytona sandboxes",
	Aliases: []string{"sandboxes"},
	GroupID: internal.SANDBOX_GROUP,
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}
type Sandbox struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func init() {
	SandboxCmd.AddCommand(ListCmd)
	SandboxCmd.AddCommand(CreateCmd)
	SandboxCmd.AddCommand(InfoCmd)
	SandboxCmd.AddCommand(DeleteCmd)
	SandboxCmd.AddCommand(StartCmd)
	SandboxCmd.AddCommand(StopCmd)
}

// add more 10 sandboxes
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List sandboxes",
		RunE: func(cmd *cobra.Command, args []string) error {
			page, _ := cmd.Flags().GetInt("page")
			limit, _ := cmd.Flags().GetInt("limit")

			// Validate input
			if page < 1 {
				return fmt.Errorf("page must be a positive integer")
			}
			if limit < 1 {
				return fmt.Errorf("limit must be a positive integer")
			}

			// Get paginated sandboxes from API
			sandboxes, total, err := apiclient.Client{}.SandboxList(page, limit)

			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
			}

			if err != nil {
				return err
			}

			// Calculate display range
			start := (page-1)*limit + 1
			end := (page-1)*limit + len(sandboxes)
			if len(sandboxes) == 0 {
				start = 0
				end = 0
			}

			// Format output
			fmt.Printf("Sandboxes (Page %d, Showing %d-%d of %d):\n",
				page, start, end, total)
			for _, sb := range sandboxes {
				fmt.Printf("ID: %s, Name: %s\n", sb.ID, sb.Name)
			}
			return nil
		},
	}

	cmd.Flags().IntP("page", "p", 1, "Page number")
	cmd.Flags().IntP("limit", "l", 10, "Items per page")
	return cmd
}
