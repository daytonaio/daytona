// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/kardianos/service"
)

const serviceName = "DaytonaServerDaemon"

type program struct {
	service.Interface
}

func Start(logFilePath string) error {
	cfg, err := getServiceConfig()
	if err != nil {
		return err
	}

	s, err := service.New(program{}, cfg)
	if err != nil {
		return err
	}

	serviceFilePath, err := getServiceFilePath(cfg)
	if err != nil {
		return err
	}

	_, err = os.Stat(serviceFilePath)
	if os.IsNotExist(err) {
		err = s.Install()
		if err != nil {
			return err
		}
	}

	logFile, err := os.OpenFile(logFilePath, os.O_TRUNC|os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	go func() {
		reader := bufio.NewReader(logFile)
		for {
			bytes := make([]byte, 1024)
			_, err := reader.Read(bytes)
			if err == nil {
				fmt.Print(string(bytes))
			}
		}
	}()

	err = s.Start()
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	status, err := s.Status()
	if err != nil {
		return err
	}

	if status == service.StatusRunning {
		return nil
	}

	err = Stop()
	if err != nil {
		return err
	}

	if status == service.StatusStopped {
		return fmt.Errorf("daemon stopped unexpectedly")
	} else {
		return fmt.Errorf("daemon status unknown")
	}
}

func Stop() error {
	cfg, err := getServiceConfig()
	if err != nil {
		return err
	}
	s, err := service.New(program{}, cfg)
	if err != nil {
		return err
	}

	err = s.Stop()
	if err != nil {
		return err
	}

	return s.Uninstall()
}

func getServiceConfig() (*service.Config, error) {
	user, ok := os.LookupEnv("USER")
	if !ok {
		return nil, fmt.Errorf("could not determine user")
	}

	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: "Daytona Server",
		Description: "Daytona Server daemon.",
		Arguments:   []string{"serve"},
	}

	switch runtime.GOOS {
	case "windows":
		return nil, fmt.Errorf("daemon mode not supported on Windows")
	case "linux":
		// Fix for running as root on Linux
		if user == "root" {
			svcConfig.UserName = user
		}
		if !strings.HasSuffix(service.Platform(), "systemd") {
			return nil, fmt.Errorf("on Linux, `server` is only supported with systemd. %s detected", service.Platform())
		}
		fallthrough
	case "darwin":
		if user != "" && user != "root" {
			svcConfig.Option = service.KeyValue{"UserService": true}
		}
	}
	svcConfig.EnvVars = util.GetEnvVarsFromShell()

	return svcConfig, nil
}

func getServiceFilePath(cfg *service.Config) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		if cfg.UserName == "root" {
			return fmt.Sprintf("/etc/systemd/system/%s.service", cfg.Name), nil
		}
		return fmt.Sprintf("%s/.config/systemd/user/%s.service", homeDir, cfg.Name), nil
	case "darwin":
		if cfg.UserName == "root" {
			return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", cfg.Name), nil
		}
		return fmt.Sprintf("%s/Library/LaunchAgents/%s.plist", homeDir, cfg.Name), nil
	}

	return "", fmt.Errorf("daemon mode not supported on current OS")
}
