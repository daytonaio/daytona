// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/os"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/pkg/browser"
)

func OpenJetbrainsIDE(activeProfile config.Profile, ide, workspaceId, projectName string) error {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	jbIde, ok := jetbrains.GetIdes()[jetbrains.Id(ide)]
	if !ok {
		return fmt.Errorf("IDE not found")
	}

	downloadPath := filepath.ToSlash(filepath.Join("/home/daytona/.cache/JetBrains", ide))

	downloadUrl := ""

	remoteOs, err := util.GetRemoteOS(projectHostname)
	if err != nil {
		return err
	}

	switch *remoteOs {
	case os.Linux_arm64:
		downloadUrl = fmt.Sprintf(jbIde.UrlTemplates.Arm64, jbIde.Version)
	case os.Linux_64_86:
		downloadUrl = fmt.Sprintf(jbIde.UrlTemplates.Amd64, jbIde.Version)
	default:
		return fmt.Errorf("JetBrains remote IDEs are only supported on Linux.")
	}

	err = downloadJetbrainsIDE(projectHostname, downloadUrl, downloadPath)
	if err != nil {
		return err
	}

	gatewayUrl := fmt.Sprintf("jetbrains-gateway://connect#host=%s&type=ssh&deploy=false&projectPath=%s&user=daytona&port=2222&idePath=%s", projectHostname, path.Join("/workspaces", projectName), url.QueryEscape(downloadPath))

	return browser.OpenURL(gatewayUrl)
}

func downloadJetbrainsIDE(projectHostname, downloadUrl, downloadPath string) error {
	if isAlreadyDownloaded(projectHostname, downloadPath) {
		view_util.RenderInfoMessage("JetBrains IDE already downloaded. Opening...")
		return nil
	}

	view_util.RenderInfoMessage(fmt.Sprintf("Downloading the IDE into the project from %s...", downloadUrl))

	downloadIdeCmd := exec.Command("ssh", projectHostname, fmt.Sprintf("mkdir -p %s && curl -fsSL %s | tar -xz -C %s --strip-components=1", downloadPath, downloadUrl, downloadPath))
	downloadIdeCmd.Stdout = io.Writer(&util.DebugLogWriter{})
	downloadIdeCmd.Stderr = io.Writer(&util.DebugLogWriter{})

	err := downloadIdeCmd.Run()
	if err != nil {
		return err
	}

	view_util.RenderInfoMessage("IDE downloaded. Opening...")

	return nil
}

func isAlreadyDownloaded(projectHostname, downloadPath string) bool {
	statCmd := exec.Command("ssh", projectHostname, fmt.Sprintf("stat %s", downloadPath))
	err := statCmd.Run()
	return err == nil
}
