package util

import (
	"os"
	"path/filepath"
)

// GetDaytonaDir returns the Daytona configuration directory
func GetDaytonaDir() string {
  // First check if DAYTONA_DIR environment variable is set
  if dir := os.Getenv("DAYTONA_DIR"); dir != "" {
      return dir
  }

  // Fall back to default location in user's home directory
  homeDir, err := os.UserHomeDir()
  if err != nil {
      return filepath.Join(os.TempDir(), ".daytona")
  }
  return filepath.Join(homeDir, ".daytona")
}