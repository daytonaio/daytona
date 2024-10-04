// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	daytona_os "github.com/daytonaio/daytona/pkg/os"
)

// Check if the requested binary is already downloaded, if not, download it
func (s *Server) GetBinaryPath(binaryName, binaryVersion string) (string, error) {
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
	if *hostOs == binaryOs && binaryVersion == s.Version {
		executable, err := os.Executable()
		if err == nil {
			f, err := os.Open(executable)
			if err == nil {
				defer f.Close() // nolint: errcheck
				return executable, nil
			}
		}
	}

	binaryPath := filepath.Join(s.config.BinariesPath, binaryVersion, binaryName)
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	downloadUrl, err := url.JoinPath(s.config.RegistryUrl, binaryVersion, binaryName)
	if err != nil {
		return "", err
	}

	err = daytona_os.DownloadFile(downloadUrl, binaryPath)
	if err != nil {
		return "", err
	}

	return binaryPath, nil
}
