package projectconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
  Use:   "import [FILE]",
  Short: "Import project configurations from a JSON file",
  Args:  cobra.ExactArgs(1),
  RunE:  runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
  filePath := args[0]
  return importProjectConfigs(cmd.Context(), filePath)
}

func importProjectConfigs(ctx context.Context, filePath string) error {
  data, err := os.ReadFile(filePath)
  if err != nil {
      return fmt.Errorf("failed to read file: %v", err)
  }

  var configs []apiclient.ProjectConfig
  err = json.Unmarshal(data, &configs)
  if err != nil {
      return fmt.Errorf("failed to parse JSON: %v", err)
  }

  apiClient := apiclient.NewAPIClient(nil)

  for _, config := range configs {
      // Convert ProjectConfig to CreateProjectConfigDTO
      dto := &apiclient.CreateProjectConfigDTO{
          Name: config.Name,
          // Add other fields that exist in CreateProjectConfigDTO
      }

      // Try to get existing config
      existing, _, err := apiClient.ProjectConfigAPI.GetProjectConfig(ctx, config.Name).Execute()
      if err == nil && existing != nil {
          // Config exists, use SetProjectConfig
          _, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(*dto).Execute()
      } else {
          // Config doesn't exist, use SetProjectConfig (since CreateProjectConfig is undefined)
          _, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(*dto).Execute()
      }

      if err != nil {
          return fmt.Errorf("failed to import project config %s: %v", config.Name, err)
      }
      fmt.Printf("Imported project config: %s\n", config.Name)
  }

  return nil
}