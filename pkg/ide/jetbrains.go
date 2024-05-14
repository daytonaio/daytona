// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util"
	ospkg "github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/pkg/browser"
)

func OpenJetbrainsIDE(activeProfile config.Profile, ide, workspaceId, projectName string) error {
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
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

	jbIdeVersion, err := getJetbrainsVersion(jbIde.ProductCode)
	if err != nil {
		return err
	}

	switch *remoteOs {
	case ospkg.Linux_arm64:
		downloadUrl = fmt.Sprintf(jbIde.UrlTemplates.Arm64, jbIdeVersion)
	case ospkg.Linux_64_86:
		downloadUrl = fmt.Sprintf(jbIde.UrlTemplates.Amd64, jbIdeVersion)
	default:
		return fmt.Errorf("JetBrains remote IDEs are only supported on Linux.")
	}

	err = downloadJetbrainsIDE(projectHostname, downloadUrl, downloadPath)
	if err != nil {
		return err
	}

	gatewayUrl := fmt.Sprintf("jetbrains-gateway://connect#host=%s&type=ssh&deploy=false&projectPath=%s&user=daytona&port=2222&idePath=%s", projectHostname, projectDir, url.QueryEscape(downloadPath))

	return browser.OpenURL(gatewayUrl)
}

func downloadJetbrainsIDE(projectHostname, downloadUrl, downloadPath string) error {
	if isAlreadyDownloaded(projectHostname, downloadPath) {
		views.RenderInfoMessage("JetBrains IDE already downloaded. Opening...")
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

	views.RenderInfoMessage("IDE downloaded. Opening...")

	return nil
}

func isAlreadyDownloaded(projectHostname, downloadPath string) bool {
	statCmd := exec.Command("ssh", projectHostname, fmt.Sprintf("stat %s", downloadPath))
	err := statCmd.Run()
	return err == nil
}

func getJetbrainsVersion(productCode string) (string, error) {
	jetbrainsDataServicesUrl := fmt.Sprintf("https://data.services.jetbrains.com/products/releases?code=%s&type=release&latest=true&build=", productCode)
	res, err := http.Get(jetbrainsDataServicesUrl)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var result map[string][]map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	for _, v := range result {
		if len(v) > 0 {
			if version, ok := v[0]["version"].(string); ok {
				return version, nil
			}
		}
	}
	return "", fmt.Errorf("jetbrains: no version found for %s", productCode)
}
