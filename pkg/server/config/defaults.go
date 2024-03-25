// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

const defaultRegistryUrl = "https://download.daytona.io/daytona"
const defaultServerDownloadUrl = "https://download.daytona.io/daytona/install.sh"
const defaultHeadscalePort = 3001
const defaultApiPort = 3000

var us_defaultFrpsConfig = types.FRPSConfig{
	Domain:   "try-us.daytona.app",
	Port:     7000,
	Protocol: "https",
}

var eu_defaultFrpsConfig = types.FRPSConfig{
	Domain:   "try-eu.daytona.app",
	Port:     7000,
	Protocol: "https",
}

func getDefaultFRPSConfig() *types.FRPSConfig {
	frpsDomain := os.Getenv("DEFAULT_FRPS_DOMAIN")
	fprsProtocol := os.Getenv("DEFAULT_FRPS_PROTOCOL")
	frpsPort := os.Getenv("DEFAULT_FRPS_PORT")
	if frpsDomain != "" && fprsProtocol != "" && frpsPort != "" {
		port, err := parsePort(frpsPort)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default", err))
		} else {
			return &types.FRPSConfig{
				Domain:   frpsDomain,
				Port:     port,
				Protocol: fprsProtocol,
			}
		}
	} else {
		log.Info("Using default FRPS config")
	}

	// Return config which responds fastest to a ping
	usReturnChan := make(chan bool)
	euReturnChan := make(chan bool)

	go func() {
		// Ping US server
		_, _ = http.Get(fmt.Sprintf("%s://%s:%d", us_defaultFrpsConfig.Protocol, us_defaultFrpsConfig.Domain, us_defaultFrpsConfig.Port))
		usReturnChan <- true
	}()

	go func() {
		// Ping EU server
		_, _ = http.Get(fmt.Sprintf("%s://%s:%d", eu_defaultFrpsConfig.Protocol, eu_defaultFrpsConfig.Domain, eu_defaultFrpsConfig.Port))
		euReturnChan <- true
	}()

	select {
	case <-usReturnChan:
		return &us_defaultFrpsConfig
	case <-euReturnChan:
		return &eu_defaultFrpsConfig
	}
}

func getDefaultConfig() (*types.ServerConfig, error) {
	providersDir, err := getDefaultProvidersDir()
	if err != nil {
		return nil, errors.New("failed to get default providers dir")
	}

	targetsPath, err := getDefaultTargetsPath()
	if err != nil {
		return nil, errors.New("failed to get default targets path")
	}

	binariesPath, err := getDefaultBinariesPath()
	if err != nil {
		return nil, errors.New("failed to get default binaries path")
	}

	c := types.ServerConfig{
		Id:                generateUuid(),
		GitProviders:      []types.GitProvider{},
		RegistryUrl:       defaultRegistryUrl,
		ProvidersDir:      providersDir,
		ServerDownloadUrl: defaultServerDownloadUrl,
		ApiPort:           defaultApiPort,
		HeadscalePort:     defaultHeadscalePort,
		TargetsFilePath:   targetsPath,
		BinariesPath:      binariesPath,
		Frps:              getDefaultFRPSConfig(),
	}

	if os.Getenv("DEFAULT_REGISTRY_URL") != "" {
		c.RegistryUrl = os.Getenv("DEFAULT_REGISTRY_URL")
	}
	if os.Getenv("DEFAULT_SERVER_DOWNLOAD_URL") != "" {
		c.ServerDownloadUrl = os.Getenv("DEFAULT_SERVER_DOWNLOAD_URL")
	}
	if os.Getenv("DEFAULT_PROVIDERS_DIR") != "" {
		c.ProvidersDir = os.Getenv("DEFAULT_PROVIDERS_DIR")
	}
	if os.Getenv("DEFAULT_TARGETS_FILE_PATH") != "" {
		c.TargetsFilePath = os.Getenv("DEFAULT_TARGETS_FILE_PATH")
	}
	if os.Getenv("DEFAULT_BINARIES_PATH") != "" {
		c.BinariesPath = os.Getenv("DEFAULT_BINARIES_PATH")
	}
	if os.Getenv("DEFAULT_API_PORT") != "" {
		apiPort, err := parsePort(os.Getenv("DEFAULT_API_PORT"))
		if err != nil {
			log.Error(fmt.Printf("%s. Using %d", err, defaultApiPort))
		} else {
			c.ApiPort = apiPort
		}
	}
	if os.Getenv("DEFAULT_HEADSCALE_PORT") != "" {
		headscalePort, err := parsePort(os.Getenv("DEFAULT_HEADSCALE_PORT"))
		if err != nil {
			log.Error(fmt.Printf("%s. Using %d", err, defaultHeadscalePort))
		} else {
			c.HeadscalePort = headscalePort
		}
	}

	return &c, nil
}

func parsePort(port string) (uint32, error) {
	p, err := strconv.Atoi(port)
	if err != nil {
		return 0, errors.New("failed to parse port")
	}
	if p < 0 || p > 65535 {
		return 0, errors.New("port out of range")
	}

	return uint32(p), nil
}

func getDefaultProvidersDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "daytona", "providers"), nil
}

func getDefaultTargetsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "targets.json"), nil
}

func getDefaultBinariesPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "binaries"), nil
}

func generateUuid() string {
	uuid := uuid.New()
	return uuid.String()
}
