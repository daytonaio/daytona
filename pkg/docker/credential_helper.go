// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type IDockerCredHelper interface {
	SetDockerConfig() error
}

type DockerCredHelper struct {
	DockerConfigFileName string
	LogWriter            io.Writer
	HomeDir              string
}

func (d *DockerCredHelper) SetDockerConfig() error {
	err := d.createDockerCredHelperExecutable()
	if err != nil {
		return err
	}

	dockerFilePath := strings.TrimSuffix(d.DockerConfigFileName, "/config.json")
	_, err = os.Stat(dockerFilePath)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(dockerFilePath, 0755)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(d.DockerConfigFileName)
	if err != nil && os.IsNotExist(err) {
		_, err := os.Create(d.DockerConfigFileName)
		if err != nil {
			return err
		}
	}

	var dockerConfigContent []byte
	dockerConfigContent, err = os.ReadFile(d.DockerConfigFileName)
	if err != nil || len(dockerConfigContent) == 0 {
		dockerConfigContent = []byte("{}")
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(dockerConfigContent, &cfg); err != nil {
		return err
	}

	cfg["credsStore"] = "daytona"

	updatedConfigContent, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(d.DockerConfigFileName, updatedConfigContent, 0755)
}

func (d *DockerCredHelper) createDockerCredHelperExecutable() error {
	content := "#!/bin/bash\ndaytona docker-cred\n"
	fileName := "docker-credential-daytona"
	filePath := "/usr/local/bin"

	_, err := os.Create(filepath.Join(filePath, fileName))
	if err != nil {
		filePath = filepath.Join(d.HomeDir, ".local", "bin")

		_, ok := os.Stat(filePath)
		if os.IsNotExist(ok) {
			err := os.MkdirAll(filePath, 0755)
			if err != nil {
				return err
			}
		}

		_, err = os.Create(filepath.Join(filePath, fileName))
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filepath.Join(filePath, fileName), []byte(content), 0755)
	if err != nil {
		return err
	}

	return os.Chmod(filepath.Join(filePath, fileName), 0755)
}
