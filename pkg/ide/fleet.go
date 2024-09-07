package ide

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	ospkg "github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/pkg/browser"
)

func OpenFleet(activeProfile config.Profile, workspaceId, projectName string) error {

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	downloadPath := filepath.ToSlash(filepath.Join("/home/daytona/.cache/Fleet"))

	downloadUrl, err := getDownloadUrl(projectHostname)
	if err != nil {
		return err
	}

	err = downloadFleet(projectHostname, downloadUrl, downloadPath)
	if err != nil {
		return err
	}

	workspaceUrl, err := launchWorkspace(projectDir)
	if err != nil {
		return err
	}

	views.RenderInfoMessage("IDE Opening...")

	return browser.OpenURL(workspaceUrl)

}

func downloadFleet(projectHostname, downloadUrl, downloadPath string) error {
	if isAlreadyDownloaded(projectHostname, downloadPath) {
		views.RenderInfoMessage("JetBrains Fleet IDE already downloaded. Opening...")
		return nil
	}

	views.RenderInfoMessage(fmt.Sprintf("Downloading the IDE into the project from %s...", downloadUrl))

	downloadIdeCmd := exec.Command("ssh", projectHostname, fmt.Sprintf("mkdir -p %s && wget -q --show-progress --progress=bar:force -pO- %s | tar -xzC %s --strip-components=1", downloadPath, downloadUrl, downloadPath))
	downloadIdeCmd.Stdout = os.Stdout
	downloadIdeCmd.Stderr = os.Stderr

	err := downloadIdeCmd.Run()
	if err != nil {
		return err
	}

	views.RenderInfoMessage("IDE downloaded...")

	return nil

}

func launchWorkspace(projectDir string) (string, error) {
	views.RenderInfoMessage("Launching workspace...")

	var stdout bytes.Buffer

	launchCommand := exec.Command("./fleet launch workspace --", "--auth=accept-everyone", "--publish", "--enableSmartMode", fmt.Sprintf("--projectDir=%s", projectDir))
	launchCommand.Stdout = &stdout
	launchCommand.Stderr = os.Stderr

	err := launchCommand.Run()
	if err != nil {
		return "", err
	}

	output := stdout.String()
	urlStart := strings.Index(output, "https://fleet.jetbrains.com")
	urlEnd := strings.Index(output[urlStart:], " ")
	if urlStart != -1 && urlEnd != -1 {
		url := output[urlStart : urlStart+urlEnd]
		fmt.Printf("Extracted URL: %s\n", url)
		return url, err
	}

	return "", fmt.Errorf("no url found in output")
}

func getDownloadUrl(projectHostname string) (string, error) {

	remoteOs, err := util.GetRemoteOS(projectHostname)
	if err != nil {
		return "", err
	}

	switch *remoteOs {
	case ospkg.Linux_64_86:
		return "https://download.jetbrains.com/product?code=FLL&release.type=preview&release.type=eap&platform=linux_x64", nil
	case ospkg.Linux_arm64:
		return "https://download.jetbrains.com/product?code=FLL&release.type=preview&release.type=eap&platform=linux_aarch64", nil
	default:
		return "", fmt.Errorf("JetBrains fleet IDE are only supported on Linux")
	}

}
