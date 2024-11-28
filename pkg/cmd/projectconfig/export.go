package projectconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
  Use:   "export [CONFIG_NAME]",
  Short: "Export project configurations to a JSON file",
  Args:  cobra.MaximumNArgs(1),
  RunE:  runExport,
}

func runExport(cmd *cobra.Command, args []string) error {
  var configName string
  if len(args) > 0 {
      configName = args[0]
  }
  return exportProjectConfigs(cmd.Context(), configName)
}

func exportProjectConfigs(ctx context.Context, configName string) error {
  if ctx == nil {
      return fmt.Errorf("context cannot be nil")
  }

  apiClient := apiclient.NewAPIClient(nil)
  if apiClient == nil {
      return fmt.Errorf("failed to create API client")
  }

  configs, err := fetchConfigs(ctx, apiClient, configName)
  if err != nil {
      return err
  }

  data, err := json.MarshalIndent(configs, "", "  ")
  if err != nil {
      return fmt.Errorf("failed to marshal JSON: %v", err)
  }

  exportDir := filepath.Join(util.GetDaytonaDir(), "exports", "project-configs")
  if err := os.MkdirAll(exportDir, 0755); err != nil {
      return fmt.Errorf("failed to create export directory: %v", err)
  }

  fileName := "project-configs.json"
  if configName != "" {
      fileName = fmt.Sprintf("%s.json", configName)
  }
  filePath := filepath.Join(exportDir, fileName)

  if err := os.WriteFile(filePath, data, 0644); err != nil {
      return fmt.Errorf("failed to write file: %v", err)
  }

  fmt.Printf("Successfully exported project configs to: %s\n", filePath)
  return nil
}

func fetchConfigs(ctx context.Context, apiClient *apiclient.APIClient, configName string) ([]apiclient.ProjectConfig, error) {
  var configs []apiclient.ProjectConfig

  if configName != "" {
      config, _, err := apiClient.ProjectConfigAPI.GetProjectConfig(ctx, configName).Execute()
      if err != nil {
          return nil, fmt.Errorf("failed to fetch project config %s: %v", configName, err)
      }
      configs = append(configs, *config)
  } else {
      configList, _, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
      if err != nil {
          return nil, fmt.Errorf("failed to fetch project configs: %v", err)
      }
      configs = configList
  }

  return configs, nil
}