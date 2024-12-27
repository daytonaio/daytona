// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build/devcontainer"

	log "github.com/sirupsen/logrus"
)

func OpenVSCode(activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string, gpgKey string) error {
	CheckAndAlertVSCodeInstalled()
	err := installRemoteSSHExtension("code")
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgKey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, projectDir)

	vscCommand := exec.Command("code", "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = vscCommand.Run()
	if err != nil {
		return err
	}

	if projectProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(projectHostname, projectProviderMetadata, devcontainer.Vscode, "*/.vscode-server/*/bin/code-server", "$HOME/.vscode-server/data/Machine/settings.json", ".daytona-customizations-lock-vscode")
}

func setupVSCodeCustomizations(projectHostname string, projectProviderMetadata string, tool devcontainer.Tool, codeServerPath string, settingsPath string, lockFileName string) error {
	// Check if customizations are already set up
	err := exec.Command("ssh", projectHostname, "test", "-f", fmt.Sprintf("$HOME/%s-%s", lockFileName, string(tool))).Run()
	if err == nil {
		return nil
	}

	fmt.Println("Setting up IDE customizations...")

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(projectProviderMetadata), &metadata); err != nil {
		return err
	}

	if devcontainerMetadata, ok := metadata["devcontainer.metadata"]; ok {
		var configs []devcontainer.Configuration
		if err := json.Unmarshal([]byte(devcontainerMetadata.(string)), &configs); err != nil {
			// Metadata can sometimes be a single object
			var config devcontainer.Configuration
			if err := json.Unmarshal([]byte(devcontainerMetadata.(string)), &config); err != nil {
				return err
			}
			configs = append(configs, config)
		}

		customizations := []devcontainer.Customizations{}

		for _, config := range configs {
			if config.Customizations != nil {
				c := config.GetCustomizations(tool)
				if c != nil {
					customizations = append(customizations, *c)
				}
			}
		}

		mergedCustomizations := devcontainer.MergeCustomizations(customizations)

		var vscodePath []byte

		fmt.Println("Waiting for code server to install...")
		for {
			time.Sleep(2 * time.Second)
			// Wait for code to be installed
			var err error
			if vscodePath, err = exec.Command("ssh", projectHostname, "find", "$HOME", "-path", fmt.Sprintf(`"%s"`, codeServerPath)).Output(); err == nil && len(vscodePath) > 0 {
				break
			}
		}

		if mergedCustomizations != nil && len(mergedCustomizations.Extensions) > 0 {
			extensionArgs := []string{}
			for _, extension := range mergedCustomizations.Extensions {
				extensionArgs = append(extensionArgs, "--install-extension", extension)
			}

			args := []string{
				projectHostname,
				strings.TrimRight(string(vscodePath), "\n"),
				"--accept-server-license-terms",
			}

			args = append(args, extensionArgs...)

			installCmd := exec.Command("ssh", args...)
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			err := installCmd.Run()
			if err != nil {
				log.Errorf("Failed to install extensions: %s", err)
			}
		}

		err := setupVSCodeSettings(projectHostname, mergedCustomizations, settingsPath)
		if err != nil {
			log.Errorf("Failed to set IDE settings: %s", err)
		}
	}

	// Create lock file to indicate that customizations are set up
	err = exec.Command("ssh", projectHostname, "touch", fmt.Sprintf("$HOME/%s-%s", lockFileName, string(tool))).Run()
	if err != nil {
		return err
	}

	fmt.Println("IDE customizations set up successfully")
	return nil
}

func setupVSCodeSettings(projectHostname string, customizations *devcontainer.Customizations, settingsPath string) error {
	if customizations == nil {
		return nil
	}

	content, err := exec.Command("ssh", projectHostname, "cat", settingsPath).Output()
	if err != nil {
		content = []byte("{}")
	}

	settings := map[string]interface{}{}
	err = json.Unmarshal(content, &settings)

	if err != nil {
		return err
	}

	if customizations.Settings != nil {
		for key, value := range customizations.Settings {
			if _, ok := settings[key]; !ok {
				settings[key] = value
			}
		}
	}

	fmt.Println("Setting up IDE settings...")

	settingsJson, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	err = exec.Command("ssh", projectHostname, "echo", fmt.Sprintf(`'%s'`, string(settingsJson)), ">", settingsPath).Run()
	if err != nil {
		return err
	}

	fmt.Println("IDE settings set up successfully")
	return nil
}

func CheckAndAlertVSCodeInstalled() {
	if err := isVSCodeInstalled(); err != nil {
		redBold := "\033[1;31m" // ANSI escape code for red and bold
		reset := "\033[0m"      // ANSI escape code to reset text formatting

		errorMessage := "Please install Visual Studio Code and ensure it's in your PATH. "
		infoMessage := "More information on: 'https://code.visualstudio.com/docs/editor/command-line#_launching-from-command-line'"
		if runtime.GOOS == "darwin" {
			infoMessage = "More information on: 'https://code.visualstudio.com/docs/setup/mac#_launching-from-the-command-line'"
		}

		log.Error(redBold + errorMessage + reset + infoMessage)

		return
	}
}

func isVSCodeInstalled() error {
	_, err := exec.LookPath("code")
	return err
}

func installRemoteSSHExtension(binaryPath string) error {
	output, err := exec.Command(binaryPath, "--list-extensions").Output()
	if err != nil {
		return err
	}

	if !strings.Contains(string(output), "ms-vscode-remote.remote-ssh") {
		fmt.Println("Installing Remote SSH extension...")
		err = exec.Command(binaryPath, "--install-extension", "ms-vscode-remote.remote-ssh").Run()
		if err != nil {
			return err
		}
		fmt.Println("Remote SSH extension successfully installed")
	}
	return nil
}
