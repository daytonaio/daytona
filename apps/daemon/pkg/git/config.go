// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"gopkg.in/ini.v1"
)

func (s *Service) SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error {
	gitConfigFileName := s.GitConfigFileName

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

	if !cfg.HasSection("safe") {
		_, err := cfg.NewSection("safe")
		if err != nil {
			return err
		}
	}
	_, err = cfg.Section("safe").NewKey("directory", s.ProjectDir)
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

		_, err := cfg.Section("user").NewKey("name", userData.Name)
		if err != nil {
			return err
		}

		_, err = cfg.Section("user").NewKey("email", userData.Email)
		if err != nil {
			return err
		}
	}

	if err := s.setSigningConfig(cfg, providerConfig, userData); err != nil {
		return err
	}

	var buf bytes.Buffer
	_, err = cfg.WriteTo(&buf)
	if err != nil {
		return err
	}

	return os.WriteFile(gitConfigFileName, buf.Bytes(), 0644)
}

func (s *Service) setSigningConfig(cfg *ini.File, providerConfig *gitprovider.GitProviderConfig, userData *gitprovider.GitUser) error {
	if providerConfig == nil || providerConfig.SigningMethod == nil || providerConfig.SigningKey == nil {
		return nil
	}

	if !cfg.HasSection("user") {
		_, err := cfg.NewSection("user")
		if err != nil {
			return err
		}
	}

	_, err := cfg.Section("user").NewKey("signingkey", *providerConfig.SigningKey)
	if err != nil {
		return err
	}

	if !cfg.HasSection("commit") {
		_, err := cfg.NewSection("commit")
		if err != nil {
			return err
		}
	}

	switch *providerConfig.SigningMethod {
	case gitprovider.SigningMethodGPG:
		_, err := cfg.Section("commit").NewKey("gpgSign", "true")
		if err != nil {
			return err
		}
	case gitprovider.SigningMethodSSH:
		err := s.configureAllowedSigners(userData.Email, *providerConfig.SigningKey)
		if err != nil {
			return err
		}

		if !cfg.HasSection("gpg") {
			_, err := cfg.NewSection("gpg")
			if err != nil {
				return err
			}
		}
		_, err = cfg.Section("gpg").NewKey("format", "ssh")
		if err != nil {
			return err
		}

		if !cfg.HasSection("gpg \"ssh\"") {
			_, err := cfg.NewSection("gpg \"ssh\"")
			if err != nil {
				return err
			}
		}

		allowedSignersFile := filepath.Join(os.Getenv("HOME"), ".ssh/allowed_signers")
		_, err = cfg.Section("gpg \"ssh\"").NewKey("allowedSignersFile", allowedSignersFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) configureAllowedSigners(email, sshKey string) error {
	homeDir := os.Getenv("HOME")
	sshDir := filepath.Join(homeDir, ".ssh")
	allowedSignersFile := filepath.Join(sshDir, "allowed_signers")

	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	entry := fmt.Sprintf("%s namespaces=\"git\" %s\n", email, sshKey)

	existingContent, err := os.ReadFile(allowedSignersFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read allowed_signers file: %w", err)
	}

	newContent := string(existingContent) + entry

	err = os.WriteFile(allowedSignersFile, []byte(newContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write to allowed_signers file: %w", err)
	}

	return nil
}
