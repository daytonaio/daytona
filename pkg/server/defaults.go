// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

const defaultRegistryUrl = "https://download.daytona.io/daytona"
const defaultServerDownloadUrl = "https://download.daytona.io/daytona/install.sh"
const defaultSamplesIndexUrl = "https://raw.githubusercontent.com/daytonaio/daytona/main/hack/samples/index.json"
const defaultHeadscalePort = 3987
const defaultApiPort = 3986
const defaultBuilderImage = "daytonaio/workspace-project:latest"
const defaultProjectImage = "daytonaio/workspace-project:latest"
const defaultProjectUser = "daytona"

const defaultLocalBuilderRegistryPort = 3988
const defaultLocalBuilderRegistryImage = "registry:2.8.3"
const defaultBuilderRegistryServer = "local"
const defaultBuildImageNamespace = ""

var defaultLogFileConfig = LogFileConfig{
	MaxSize:    100, // megabytes
	MaxBackups: 7,
	MaxAge:     15, // days
	LocalTime:  true,
	Compress:   true,
}

var us_defaultFrpsConfig = FRPSConfig{
	Domain:   "try-us.daytona.app",
	Port:     7000,
	Protocol: "https",
}

var eu_defaultFrpsConfig = FRPSConfig{
	Domain:   "try-eu.daytona.app",
	Port:     7000,
	Protocol: "https",
}

func getDefaultFRPSConfig() *FRPSConfig {
	frpsDomain := os.Getenv("DEFAULT_FRPS_DOMAIN")
	fprsProtocol := os.Getenv("DEFAULT_FRPS_PROTOCOL")
	frpsPort := os.Getenv("DEFAULT_FRPS_PORT")
	if frpsDomain != "" && fprsProtocol != "" && frpsPort != "" {
		port, err := parsePort(frpsPort)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default", err))
		} else {
			return &FRPSConfig{
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

func getDefaultLogFileConfig() *LogFileConfig {
	logFilePath, err := getDefaultLogFilePath()
	if err != nil {
		log.Error("failed to get default log file path")
	}

	logFileConfig := LogFileConfig{
		Path:       logFilePath,
		MaxSize:    defaultLogFileConfig.MaxSize,
		MaxBackups: defaultLogFileConfig.MaxBackups,
		MaxAge:     defaultLogFileConfig.MaxAge,
		LocalTime:  defaultLogFileConfig.LocalTime,
		Compress:   defaultLogFileConfig.Compress,
	}

	logFileMaxSize := os.Getenv("DEFAULT_LOG_FILE_MAX_SIZE")
	if logFileMaxSize != "" {
		value, err := strconv.Atoi(logFileMaxSize)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file max size.", err))
		} else {
			logFileConfig.MaxSize = value
		}
	}

	logFileMaxBackups := os.Getenv("DEFAULT_LOG_FILE_MAX_BACKUPS")
	if logFileMaxBackups != "" {
		value, err := strconv.Atoi(logFileMaxBackups)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file max backups.", err))
		} else {
			logFileConfig.MaxBackups = value
		}
	}

	logFileMaxAge := os.Getenv("DEFAULT_LOG_FILE_MAX_AGE")
	if logFileMaxAge != "" {
		value, err := strconv.Atoi(logFileMaxAge)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file max age.", err))
		} else {
			logFileConfig.MaxAge = value
		}
	}

	logFileLocalTime := os.Getenv("DEFAULT_LOG_FILE_LOCAL_TIME")
	if logFileLocalTime != "" {
		value, err := strconv.ParseBool(logFileLocalTime)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file local time.", err))
		} else {
			logFileConfig.LocalTime = value
		}
	}

	logFileCompress := os.Getenv("DEFAULT_LOG_FILE_COMPRESS")
	if logFileCompress != "" {
		value, err := strconv.ParseBool(logFileCompress)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file compress.", err))
		} else {
			logFileConfig.Compress = value
		}
	}

	return &logFileConfig
}

func getDefaultConfig() (*Config, error) {
	providersDir, err := getDefaultProvidersDir()
	if err != nil {
		return nil, errors.New("failed to get default providers dir")
	}

	binariesPath, err := getDefaultBinariesPath()
	if err != nil {
		return nil, errors.New("failed to get default binaries path")
	}

	c := Config{
		Id:                        uuid.NewString(),
		RegistryUrl:               defaultRegistryUrl,
		ProvidersDir:              providersDir,
		ServerDownloadUrl:         defaultServerDownloadUrl,
		ApiPort:                   defaultApiPort,
		HeadscalePort:             defaultHeadscalePort,
		BinariesPath:              binariesPath,
		Frps:                      getDefaultFRPSConfig(),
		LogFile:                   getDefaultLogFileConfig(),
		DefaultProjectImage:       defaultProjectImage,
		DefaultProjectUser:        defaultProjectUser,
		BuilderImage:              defaultBuilderImage,
		LocalBuilderRegistryPort:  defaultLocalBuilderRegistryPort,
		LocalBuilderRegistryImage: defaultLocalBuilderRegistryImage,
		BuilderRegistryServer:     defaultBuilderRegistryServer,
		BuildImageNamespace:       defaultBuildImageNamespace,
		SamplesIndexUrl:           defaultSamplesIndexUrl,
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
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "providers"), nil
}

func getDefaultLogFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "daytona.log"), nil
}

func getDefaultBinariesPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "binaries"), nil
}
