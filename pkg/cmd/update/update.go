package update

import (
	"context"
	"fmt"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

var (
  version        string
  currentVersion = "0.0.1" // Replace this with your actual version or use ldflags to inject it dynamically
)

// NewUpdateCmd creates the `update` command for the Daytona CLI.
func NewUpdateCmd() *cobra.Command {
  cmd := &cobra.Command{
  	Use:   "update",
  	Short: "Update Daytona CLI to the latest version",
  	Long:  "Update Daytona CLI to the latest version or a specified version",
  	RunE:  runUpdate,
  }

  cmd.Flags().StringVarP(&version, "version", "v", "", "Specify a version to update to")

  return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
  ctx := context.Background()

  // Define the GitHub repository slug
  repo := selfupdate.ParseSlug("daytonaio/daytona")

  var latest *selfupdate.Release
  var err error

  if version == "" {
  	// Detect the latest version
  	var found bool
  	latest, found, err = selfupdate.DetectLatest(ctx, repo)
  	if err != nil {
  		return fmt.Errorf("error occurred while detecting version: %w", err)
  	}
  	if !found {
  		return fmt.Errorf("latest version for %s/%s could not be found from GitHub repository", runtime.GOOS, runtime.GOARCH)
  	}

  	// Check if the current version is already the latest
  	if latest.LessOrEqual(currentVersion) {
  		fmt.Printf("Current version (%s) is the latest\n", currentVersion)
  		return nil
  	}

  	version = latest.Version()
  }

  fmt.Printf("Updating to version %s...\n", version)

  // Get the path to the current executable
  exe, err := selfupdate.ExecutablePath()
  if err != nil {
  	return fmt.Errorf("could not locate executable path: %w", err)
  }

  // Create an updater with default configuration
  updater, err := selfupdate.NewUpdater(selfupdate.Config{})
  if err != nil {
  	return fmt.Errorf("failed to create updater: %w", err)
  }

  // Perform the update
  if err := updater.UpdateTo(ctx, latest, exe); err != nil {
  	return fmt.Errorf("error occurred while updating binary: %w", err)
  }

  fmt.Printf("Successfully updated to version %s\n", version)
  fmt.Printf("Changelog: https://github.com/daytonaio/daytona/releases/tag/v%s\n", version)

  return nil
}