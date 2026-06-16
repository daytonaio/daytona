// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitconfig "github.com/go-git/go-git/v5/plumbing/format/config"
)

// SetConfigValue sets a git config key/value at the given scope ("global"
// (default), "local" or "system"). It uses go-git's low-level config codec, not
// repo.SetConfig, whose marshaller clobbers arbitrary keys in known sections.
func (s *Service) SetConfigValue(key, value, scope string) error {
	section, subsection, name, err := splitConfigKey(key)
	if err != nil {
		return err
	}

	path, err := s.configPathForScope(scope)
	if err != nil {
		return err
	}

	cfg, err := decodeConfigFile(path)
	if err != nil {
		return err
	}

	if subsection == "" {
		cfg.Section(section).SetOption(name, value)
	} else {
		cfg.Section(section).Subsection(subsection).SetOption(name, value)
	}

	return encodeConfigFile(path, cfg)
}

// GetConfigValue returns the value for key at the given scope, or nil when unset.
func (s *Service) GetConfigValue(key, scope string) (*string, error) {
	section, subsection, name, err := splitConfigKey(key)
	if err != nil {
		return nil, err
	}

	path, err := s.configPathForScope(scope)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	cfg, err := decodeConfigFile(path)
	if err != nil {
		return nil, err
	}

	if subsection == "" {
		sec := cfg.Section(section)
		if !sec.HasOption(name) {
			return nil, nil
		}
		value := sec.Option(name)
		return &value, nil
	}

	sub := cfg.Section(section).Subsection(subsection)
	if !sub.HasOption(name) {
		return nil, nil
	}
	value := sub.Option(name)
	return &value, nil
}

func (s *Service) ConfigureUser(name, email, scope string) error {
	if err := s.SetConfigValue("user.name", name, scope); err != nil {
		return err
	}
	return s.SetConfigValue("user.email", email, scope)
}

func (s *Service) configPathForScope(scope string) (string, error) {
	switch normalizeScope(scope) {
	case "local":
		if s.WorkDir == "" {
			return "", fmt.Errorf("local config scope requires a repository path")
		}
		// Non-bare repos keep config under .git/; bare repos keep it at the repo root.
		if info, statErr := os.Stat(filepath.Join(s.WorkDir, ".git")); statErr == nil && info.IsDir() {
			return filepath.Join(s.WorkDir, ".git", "config"), nil
		}
		return filepath.Join(s.WorkDir, "config"), nil
	case "global":
		if override := os.Getenv("GIT_CONFIG_GLOBAL"); override != "" {
			return override, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		gitconfigPath := filepath.Join(home, ".gitconfig")
		// Prefer an existing ~/.gitconfig, then the XDG path, else create ~/.gitconfig.
		if _, err := os.Stat(gitconfigPath); err == nil {
			return gitconfigPath, nil
		}
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg == "" {
			xdg = filepath.Join(home, ".config")
		}
		if xdgPath := filepath.Join(xdg, "git", "config"); fileExists(xdgPath) {
			return xdgPath, nil
		}
		return gitconfigPath, nil
	case "system":
		if override := os.Getenv("GIT_CONFIG_SYSTEM"); override != "" {
			return override, nil
		}
		return "/etc/gitconfig", nil
	default:
		return "", fmt.Errorf("unsupported config scope %q (supported: global, local, system)", scope)
	}
}

func decodeConfigFile(path string) (*gitconfig.Config, error) {
	cfg := gitconfig.New()
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if len(content) == 0 {
		return cfg, nil
	}
	if err := gitconfig.NewDecoder(bytes.NewReader(content)).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func encodeConfigFile(path string, cfg *gitconfig.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := gitconfig.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeScope(scope string) string {
	if scope == "" {
		return "global"
	}
	return scope
}

// splitConfigKey splits "user.name" -> ("user", "", "name") and
// "remote.origin.url" -> ("remote", "origin", "url").
func splitConfigKey(key string) (section, subsection, name string, err error) {
	first := strings.Index(key, ".")
	last := strings.LastIndex(key, ".")
	if first <= 0 || last == len(key)-1 {
		return "", "", "", fmt.Errorf("invalid config key %q (expected section.name or section.subsection.name)", key)
	}

	if first == last {
		return key[:first], "", key[first+1:], nil
	}

	return key[:first], key[first+1 : last], key[last+1:], nil
}
