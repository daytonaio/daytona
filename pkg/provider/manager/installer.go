// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	goos "os"
	"path/filepath"
	"runtime"

	"github.com/daytonaio/daytona/pkg/os"
	log "github.com/sirupsen/logrus"
)

func (m *ProviderManager) GetProvidersManifest() (*ProvidersManifest, error) {
	manifestUrl := fmt.Sprintf("%s/providers/manifest.json", m.registryUrl)

	resp, err := http.Get(manifestUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	manifestJson, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest ProvidersManifest
	err = json.Unmarshal(manifestJson, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (m *ProviderManager) DownloadProvider(ctx context.Context, downloadUrls map[os.OperatingSystem]string, providerName string) (string, error) {
	downloadPath := filepath.Join(m.baseDir, providerName, providerName)
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	if _, err := goos.Stat(downloadPath); err == nil {
		return "", providerAlreadyDownloadedError(providerName)
	}

	log.Info("Downloading " + providerName)

	operatingSystem, err := os.GetOperatingSystem()
	if err != nil {
		return "", err
	}

	err = os.DownloadFile(ctx, downloadUrls[*operatingSystem], downloadPath)
	if err != nil {
		return "", err
	}

	return downloadPath, nil
}
