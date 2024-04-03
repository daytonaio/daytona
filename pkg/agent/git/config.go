// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"os"
	"path"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"gopkg.in/ini.v1"
)

func SetGitConfig(userData *serverapiclient.GitUserData) error {
	gitConfigFileName := path.Join(os.Getenv("HOME"), ".gitconfig")

	var gitConfigContent []byte
	gitConfigContent, err := os.ReadFile(gitConfigFileName)
	if err != nil {
		gitConfigContent = []byte{}
	}

	cfg, err := ini.Load(gitConfigContent)
	if err != nil {
		return err
	}

	if !cfg.HasSection("credential") {
		_, err := cfg.NewSection("credential")
		if err != nil {
			return err
		}
	}

	_, err = cfg.Section("credential").NewKey("helper", "/usr/local/bin/daytona git-cred")
	if err != nil {
		return err
	}

	if userData != nil {
		if !cfg.HasSection("user") {
			_, err := cfg.NewSection("user")
			if err != nil {
				return err
			}
		}

		if userData.Name != nil {
			_, err := cfg.Section("user").NewKey("name", *userData.Name)
			if err != nil {
				return err
			}
		}
		if userData.Email != nil {
			_, err := cfg.Section("user").NewKey("email", *userData.Email)
			if err != nil {
				return err
			}
		}
	}

	var buf bytes.Buffer
	_, err = cfg.WriteTo(&buf)
	if err != nil {
		return err
	}

	err = os.WriteFile(gitConfigFileName, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
