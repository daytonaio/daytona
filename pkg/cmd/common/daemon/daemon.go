// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/kardianos/service"
)

type program struct {
	service.Interface
}

var ErrDaemonNotInstalled = errors.New("daemon not installed")

func Start(logFilePath string, svcConfig *service.Config) error {
	cfg, err := getServiceConfig(svcConfig)
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

	err = Stop(svcConfig)
	if err != nil {
		return err
	}

	if status == service.StatusStopped {
		return errors.New("daemon stopped unexpectedly")
	} else {
		return errors.New("daemon status unknown")
	}
}

func Stop(svcConfig *service.Config) error {
	cfg, err := getServiceConfig(svcConfig)
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
	if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
		return ErrDaemonNotInstalled
	}

	err = s.Stop()
	if err != nil {
		return err
	}

	return s.Uninstall()
}

func getServiceConfig(svcConfig *service.Config) (*service.Config, error) {
	if runtime.GOOS == "windows" {
		return nil, errors.New("daemon mode not supported on Windows")
	}

	user, ok := os.LookupEnv("USER")
	if !ok {
		return nil, errors.New("could not determine user")
	}

	switch runtime.GOOS {
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

	return "", errors.New("daemon mode not supported on current OS")
}
