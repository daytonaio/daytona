// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

const spinnerThreshold = 10

// deleteDryRunItem describes one sandbox that would be deleted.
type deleteDryRunItem struct {
	Id    string `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	State string `json:"state" yaml:"state"`
}

// deleteDryRunResult is the structured output of `delete --dry-run`.
type deleteDryRunResult struct {
	DryRun    bool               `json:"dryRun" yaml:"dryRun"`
	Count     int                `json:"count" yaml:"count"`
	Sandboxes []deleteDryRunItem `json:"sandboxes" yaml:"sandboxes"`
}

type deleteBulkDeletedItem struct {
	Id   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

type deleteBulkFailedItem struct {
	Id    string `json:"id" yaml:"id"`
	Error string `json:"error" yaml:"error"`
}

// deleteBulkResult is the structured output of `delete --all`. Deleted and
// Failed are always non-nil so both keys appear in the output.
type deleteBulkResult struct {
	DryRun  bool                    `json:"dryRun" yaml:"dryRun"`
	Count   int                     `json:"count" yaml:"count"`
	Deleted []deleteBulkDeletedItem `json:"deleted" yaml:"deleted"`
	Failed  []deleteBulkFailedItem  `json:"failed" yaml:"failed"`
}

// deleteSingleResult is the structured output of a single sandbox delete.
// Name is a pointer so a missing sandbox renders "name": null.
type deleteSingleResult struct {
	Id      string  `json:"id" yaml:"id"`
	Name    *string `json:"name" yaml:"name"`
	Deleted bool    `json:"deleted" yaml:"deleted"`
	Found   bool    `json:"found" yaml:"found"`
}

var DeleteCmd = &cobra.Command{
	Use:   "delete [SANDBOX_ID | SANDBOX_NAME]",
	Short: "Delete a sandbox",
	Example: `  daytona delete my-sandbox
  daytona delete my-sandbox --wait --format json
  daytona delete --all --dry-run
  daytona delete --all --yes`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if len(args) == 0 && !allFlag {
			return clierr.New(clierr.CategoryUsage, "missing required argument: sandbox ID or name (or pass --all)")
		}

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return deleteAllSandboxes(ctx, apiClient)
		}

		return deleteSingleSandbox(ctx, apiClient, args[0])
	},
}

// sandboxDryRunResult builds the dry-run payload for the given sandboxes.
func sandboxDryRunResult(items []apiclient.SandboxListItem) deleteDryRunResult {
	result := deleteDryRunResult{DryRun: true, Count: len(items), Sandboxes: []deleteDryRunItem{}}
	for _, sb := range items {
		var state string
		if sb.State != nil {
			state = string(*sb.State)
		}
		result.Sandboxes = append(result.Sandboxes, deleteDryRunItem{Id: sb.Id, Name: sb.Name, State: state})
	}
	return result
}

// newDeleteBulkResult builds an empty bulk-delete payload with non-nil
// Deleted and Failed slices.
func newDeleteBulkResult(count int) deleteBulkResult {
	return deleteBulkResult{Count: count, Deleted: []deleteBulkDeletedItem{}, Failed: []deleteBulkFailedItem{}}
}

// awaitSandboxDeleted polls the sandbox until the API reports it gone
// (not found) or its state is DESTROYED. A timeout <= 0 waits indefinitely.
func awaitSandboxDeleted(ctx context.Context, apiClient *apiclient.APIClient, id string, timeout time.Duration) error {
	var expired <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		expired = timer.C
	}

	for {
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, id).Execute()
		if err != nil {
			handled := apiclient_cli.HandleErrorResponse(res, err)
			if clierr.HasCategory(handled, clierr.CategoryNotFound) {
				return nil
			}
			return handled
		}

		if sandbox.State != nil && *sandbox.State == apiclient.SANDBOXSTATE_DESTROYED {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-expired:
			return clierr.Newf(clierr.CategoryTimeout, "timed out after %s waiting for sandbox %q to be deleted", timeout, id)
		case <-time.After(time.Second):
		}
	}
}

func deleteAllSandboxes(ctx context.Context, apiClient *apiclient.APIClient) error {
	var cursor *string
	limit := float32(200.0) // 200 is the maximum limit for the API
	var allSandboxes []apiclient.SandboxListItem

	for {
		request := apiClient.SandboxAPI.ListSandboxes(ctx).Limit(limit)
		if cursor != nil {
			request = request.Cursor(*cursor)
		}

		sandboxBatch, res, err := request.Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		allSandboxes = append(allSandboxes, sandboxBatch.Items...)

		if !sandboxBatch.NextCursor.IsSet() || sandboxBatch.NextCursor.Get() == nil {
			break
		}
		cursor = sandboxBatch.NextCursor.Get()
	}

	if deleteDryRunFlag {
		if common.FormatFlag != "" {
			common.NewFormatter(sandboxDryRunResult(allSandboxes)).Print()
			return nil
		}
		if len(allSandboxes) == 0 {
			view_common.RenderInfoMessageBold("No sandboxes to delete")
			return nil
		}
		for _, sb := range allSandboxes {
			view_common.RenderInfoMessage(fmt.Sprintf("Would delete sandbox %s (%s)", sb.Name, sb.Id))
		}
		return nil
	}

	if len(allSandboxes) == 0 {
		if common.FormatFlag != "" {
			common.NewFormatter(newDeleteBulkResult(0)).Print()
			return nil
		}
		view_common.RenderInfoMessageBold("No sandboxes to delete")
		return nil
	}

	if !deleteYesFlag {
		confirmed, err := view_common.Confirm(fmt.Sprintf("Delete %d sandboxes?", len(allSandboxes)))
		if err != nil {
			return err
		}
		if !confirmed {
			return errors.New("aborted")
		}
	}

	result := newDeleteBulkResult(len(allSandboxes))

	deleteFn := func() error {
		var wg sync.WaitGroup
		var mu sync.Mutex
		sem := make(chan struct{}, 10) // limit to 10 concurrent deletes

		for _, sb := range allSandboxes {
			wg.Add(1)
			go func(sb apiclient.SandboxListItem) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				_, res, err := apiClient.SandboxAPI.DeleteSandbox(ctx, sb.Id).Execute()
				if err != nil {
					err = apiclient_cli.HandleErrorResponse(res, err)
				} else if deleteWaitFlag {
					err = awaitSandboxDeleted(ctx, apiClient, sb.Id, deleteTimeoutFlag)
				}

				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					if common.FormatFlag == "" {
						fmt.Printf("Failed to delete sandbox %s: %s\n", sb.Id, err)
					}
					result.Failed = append(result.Failed, deleteBulkFailedItem{Id: sb.Id, Error: err.Error()})
				} else {
					result.Deleted = append(result.Deleted, deleteBulkDeletedItem{Id: sb.Id, Name: sb.Name})
				}
			}(sb)
		}
		wg.Wait()
		return nil
	}

	if common.FormatFlag == "" && len(allSandboxes) > spinnerThreshold {
		if err := views_util.WithInlineSpinner("Deleting all sandboxes", deleteFn); err != nil {
			return err
		}
	} else if err := deleteFn(); err != nil {
		return err
	}

	if common.FormatFlag != "" {
		common.NewFormatter(result).Print()
		return nil
	}

	view_common.RenderInfoMessageBold(fmt.Sprintf("Deleted %d sandboxes", len(result.Deleted)))
	return nil
}

func deleteSingleSandbox(ctx context.Context, apiClient *apiclient.APIClient, sandboxIdOrNameArg string) error {
	sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdOrNameArg).Execute()
	if err != nil {
		handled := apiclient_cli.HandleErrorResponse(res, err)
		if deleteIgnoreNotFoundFlag && clierr.HasCategory(handled, clierr.CategoryNotFound) {
			return sandboxNotFoundOutput(sandboxIdOrNameArg)
		}
		return handled
	}

	if deleteDryRunFlag {
		if common.FormatFlag != "" {
			var state string
			if sandbox.State != nil {
				state = string(*sandbox.State)
			}
			result := deleteDryRunResult{
				DryRun:    true,
				Count:     1,
				Sandboxes: []deleteDryRunItem{{Id: sandbox.Id, Name: sandbox.Name, State: state}},
			}
			common.NewFormatter(result).Print()
			return nil
		}
		view_common.RenderInfoMessage(fmt.Sprintf("Would delete sandbox %s (%s)", sandbox.Name, sandbox.Id))
		return nil
	}

	_, res, err = apiClient.SandboxAPI.DeleteSandbox(ctx, sandbox.Id).Execute()
	if err != nil {
		handled := apiclient_cli.HandleErrorResponse(res, err)
		if deleteIgnoreNotFoundFlag && clierr.HasCategory(handled, clierr.CategoryNotFound) {
			return sandboxNotFoundOutput(sandboxIdOrNameArg)
		}
		return handled
	}

	if deleteWaitFlag {
		if err := awaitSandboxDeleted(ctx, apiClient, sandbox.Id, deleteTimeoutFlag); err != nil {
			return err
		}
	}

	if common.FormatFlag != "" {
		name := sandbox.Name
		common.NewFormatter(deleteSingleResult{Id: sandbox.Id, Name: &name, Deleted: true, Found: true}).Print()
		return nil
	}

	view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s deleted", sandboxIdOrNameArg))
	return nil
}

// sandboxNotFoundOutput reports a missing sandbox as success (--ignore-not-found).
func sandboxNotFoundOutput(arg string) error {
	if common.FormatFlag != "" {
		common.NewFormatter(deleteSingleResult{Id: arg, Name: nil, Deleted: false, Found: false}).Print()
		return nil
	}
	view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s not found, nothing to delete", arg))
	return nil
}

var (
	allFlag                  bool
	deleteYesFlag            bool
	deleteDryRunFlag         bool
	deleteIgnoreNotFoundFlag bool
	deleteWaitFlag           bool
	deleteTimeoutFlag        time.Duration
)

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all sandboxes")
	DeleteCmd.Flags().BoolVarP(&deleteYesFlag, "yes", "y", false, "Skip the confirmation prompt for bulk deletes")
	DeleteCmd.Flags().BoolVar(&deleteDryRunFlag, "dry-run", false, "Show what would be deleted without deleting anything")
	DeleteCmd.Flags().BoolVar(&deleteIgnoreNotFoundFlag, "ignore-not-found", false, "Treat a missing sandbox as a successful delete")
	DeleteCmd.Flags().BoolVar(&deleteWaitFlag, "wait", false, "Wait until the sandbox is fully deleted")
	DeleteCmd.Flags().DurationVar(&deleteTimeoutFlag, "timeout", 5*time.Minute, "Maximum time to wait with --wait (0 waits indefinitely)")
	common.RegisterFormatFlag(DeleteCmd)
}
