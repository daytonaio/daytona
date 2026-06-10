// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"errors"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

// deleteDryRunItem describes one snapshot that would be deleted.
type deleteDryRunItem struct {
	Id    string `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	State string `json:"state" yaml:"state"`
}

// deleteDryRunResult is the structured output of `delete --dry-run`.
type deleteDryRunResult struct {
	DryRun    bool               `json:"dryRun" yaml:"dryRun"`
	Count     int                `json:"count" yaml:"count"`
	Snapshots []deleteDryRunItem `json:"snapshots" yaml:"snapshots"`
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

// deleteSingleResult is the structured output of a single snapshot delete.
// Name is a pointer so a missing snapshot renders "name": null.
type deleteSingleResult struct {
	Id      string  `json:"id" yaml:"id"`
	Name    *string `json:"name" yaml:"name"`
	Deleted bool    `json:"deleted" yaml:"deleted"`
	Found   bool    `json:"found" yaml:"found"`
}

var DeleteCmd = &cobra.Command{
	Use:     "delete [SNAPSHOT_ID | SNAPSHOT_NAME]",
	Short:   "Delete a snapshot",
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if allFlag {
				return deleteAllSnapshots(ctx, apiClient)
			}
			return cmd.Help()
		}

		return deleteSingleSnapshot(ctx, apiClient, args[0])
	},
}

// deleteIsNotFound reports whether err is a not_found CLI error.
func deleteIsNotFound(err error) bool {
	var cliErr *clierr.Error
	return errors.As(err, &cliErr) && cliErr.Category == clierr.CategoryNotFound
}

// snapshotDryRunResult builds the dry-run payload for the given snapshots.
func snapshotDryRunResult(items []apiclient.SnapshotDto) deleteDryRunResult {
	result := deleteDryRunResult{DryRun: true, Count: len(items), Snapshots: []deleteDryRunItem{}}
	for _, snapshot := range items {
		result.Snapshots = append(result.Snapshots, deleteDryRunItem{Id: snapshot.Id, Name: snapshot.Name, State: string(snapshot.State)})
	}
	return result
}

// newDeleteBulkResult builds an empty bulk-delete payload with non-nil
// Deleted and Failed slices.
func newDeleteBulkResult(count int) deleteBulkResult {
	return deleteBulkResult{Count: count, Deleted: []deleteBulkDeletedItem{}, Failed: []deleteBulkFailedItem{}}
}

func deleteAllSnapshots(ctx context.Context, apiClient *apiclient.APIClient) error {
	page := float32(1.0)
	limit := float32(200.0) // 200 is the maximum limit for the API
	var allSnapshots []apiclient.SnapshotDto

	for {
		snapshotBatch, res, err := apiClient.SnapshotsAPI.GetAllSnapshots(ctx).Page(page).Limit(limit).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		allSnapshots = append(allSnapshots, snapshotBatch.Items...)

		if len(snapshotBatch.Items) < int(limit) || page >= snapshotBatch.TotalPages {
			break
		}
		page++
	}

	if deleteDryRunFlag {
		if common.FormatFlag != "" {
			common.NewFormatter(snapshotDryRunResult(allSnapshots)).Print()
			return nil
		}
		if len(allSnapshots) == 0 {
			view_common.RenderInfoMessageBold("No snapshots to delete")
			return nil
		}
		for _, snapshot := range allSnapshots {
			view_common.RenderInfoMessage(fmt.Sprintf("Would delete snapshot %s (%s)", snapshot.Name, snapshot.Id))
		}
		return nil
	}

	if len(allSnapshots) == 0 {
		if common.FormatFlag != "" {
			common.NewFormatter(newDeleteBulkResult(0)).Print()
			return nil
		}
		view_common.RenderInfoMessageBold("No snapshots to delete")
		return nil
	}

	if !deleteYesFlag {
		confirmed, err := view_common.Confirm(fmt.Sprintf("Delete %d snapshots?", len(allSnapshots)))
		if err != nil {
			return err
		}
		if !confirmed {
			return errors.New("aborted")
		}
	}

	result := newDeleteBulkResult(len(allSnapshots))

	for _, snapshot := range allSnapshots {
		res, err := apiClient.SnapshotsAPI.RemoveSnapshot(ctx, snapshot.Id).Execute()
		if err != nil {
			handled := apiclient_cli.HandleErrorResponse(res, err)
			if common.FormatFlag == "" {
				fmt.Printf("Failed to delete snapshot %s: %s\n", snapshot.Id, handled)
			}
			result.Failed = append(result.Failed, deleteBulkFailedItem{Id: snapshot.Id, Error: handled.Error()})
		} else {
			result.Deleted = append(result.Deleted, deleteBulkDeletedItem{Id: snapshot.Id, Name: snapshot.Name})
		}
	}

	if common.FormatFlag != "" {
		common.NewFormatter(result).Print()
		return nil
	}

	view_common.RenderInfoMessageBold(fmt.Sprintf("Deleted %d snapshots", len(result.Deleted)))
	return nil
}

func deleteSingleSnapshot(ctx context.Context, apiClient *apiclient.APIClient, snapshotIdOrName string) error {
	snapshot, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, snapshotIdOrName).Execute()
	if err != nil {
		handled := apiclient_cli.HandleErrorResponse(res, err)
		if deleteIgnoreNotFoundFlag && deleteIsNotFound(handled) {
			return snapshotNotFoundOutput(snapshotIdOrName)
		}
		return handled
	}

	if deleteDryRunFlag {
		if common.FormatFlag != "" {
			result := deleteDryRunResult{
				DryRun:    true,
				Count:     1,
				Snapshots: []deleteDryRunItem{{Id: snapshot.Id, Name: snapshot.Name, State: string(snapshot.State)}},
			}
			common.NewFormatter(result).Print()
			return nil
		}
		view_common.RenderInfoMessage(fmt.Sprintf("Would delete snapshot %s (%s)", snapshot.Name, snapshot.Id))
		return nil
	}

	res, err = apiClient.SnapshotsAPI.RemoveSnapshot(ctx, snapshot.Id).Execute()
	if err != nil {
		handled := apiclient_cli.HandleErrorResponse(res, err)
		if deleteIgnoreNotFoundFlag && deleteIsNotFound(handled) {
			return snapshotNotFoundOutput(snapshotIdOrName)
		}
		return handled
	}

	if common.FormatFlag != "" {
		name := snapshot.Name
		common.NewFormatter(deleteSingleResult{Id: snapshot.Id, Name: &name, Deleted: true, Found: true}).Print()
		return nil
	}

	view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s deleted", snapshotIdOrName))
	return nil
}

// snapshotNotFoundOutput reports a missing snapshot as success (--ignore-not-found).
func snapshotNotFoundOutput(arg string) error {
	if common.FormatFlag != "" {
		common.NewFormatter(deleteSingleResult{Id: arg, Name: nil, Deleted: false, Found: false}).Print()
		return nil
	}
	view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s not found, nothing to delete", arg))
	return nil
}

var (
	allFlag                  bool
	deleteYesFlag            bool
	deleteDryRunFlag         bool
	deleteIgnoreNotFoundFlag bool
)

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all snapshots")
	DeleteCmd.Flags().BoolVarP(&deleteYesFlag, "yes", "y", false, "Skip the confirmation prompt for bulk deletes")
	DeleteCmd.Flags().BoolVar(&deleteDryRunFlag, "dry-run", false, "Show what would be deleted without deleting anything")
	DeleteCmd.Flags().BoolVar(&deleteIgnoreNotFoundFlag, "ignore-not-found", false, "Treat a missing snapshot as a successful delete")
	common.RegisterFormatFlag(DeleteCmd)
}
