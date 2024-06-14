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

	"github.com/kardianos/service"
)

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
	err = s.Install()
	if err != nil {
		return err
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
		Name:        "DaytonaServerDaemon",
		DisplayName: "Daytona Server",
		Description: "This is the Daytona Server daemon.",
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
			return nil, fmt.Errorf("on Linux, `server -d` is only supported with systemd. %s detected", service.Platform())
		}
		fallthrough
	case "darwin":
		if user != "" && user != "root" {
			svcConfig.Option = service.KeyValue{"UserService": true}
		}
	}

	return svcConfig, nil
}
