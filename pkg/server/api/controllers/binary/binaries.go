package binary

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/internal"
	daytona_os "github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/gin-gonic/gin"
)

// Serves the Daytona binary based on the requested version and name.
// The name of the binary follows the pattern: daytona-<os>-<arch>[.exe]
func GetBinary(ctx *gin.Context) {
	binaryVersion := ctx.Param("version")
	binaryName := ctx.Param("binaryName")

	binaryPath, err := getBinaryPath(binaryName, binaryVersion)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get binary path: %s", err.Error()))
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(binaryPath)))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(binaryPath)
}

// Check if the requested binary is already downloaded, if not, download it
func getBinaryPath(binaryName, binaryVersion string) (string, error) {
	hostOs, err := daytona_os.GetOperatingSystem()
	if err != nil {
		return "", err
	}

	var binaryOs daytona_os.OperatingSystem
	split := strings.Split(binaryName, "-")
	if len(split) != 3 {
		return "", fmt.Errorf("invalid binary name: %s", binaryName)
	}

	binaryOs = daytona_os.OperatingSystem(fmt.Sprintf("%s-%s", split[1], strings.TrimSuffix(split[2], ".exe")))

	// If the requested binary is the same as the host, return the current binary path
	if *hostOs == binaryOs && binaryVersion == internal.Version {
		return os.Executable()
	}

	c, err := config.GetConfig()
	if err != nil {
		return "", err
	}

	binaryPath := filepath.Join(c.BinariesPath, binaryVersion, binaryName)
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	downloadUrl, err := url.JoinPath(c.RegistryUrl, binaryVersion, binaryName)
	if err != nil {
		return "", err
	}

	err = daytona_os.DownloadFile(downloadUrl, binaryPath)
	if err != nil {
		return "", err
	}

	return binaryPath, nil
}
