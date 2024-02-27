// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"os"
	"path"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

type SshAuthType = string

const (
	SshAuthTypePassword   SshAuthType = "password"
	SshAuthTypePrivateKey SshAuthType = "private-key"
)

func PrivateKeyForm(authType *SshAuthType, user, privateKeyPath *string) []*huh.Group {
	// Get private key list from ~/.ssh
	privateKeyOptions := []huh.Option[string]{}

	// Regex that will find all public keys and config files
	// Used below to find private keys
	regex := regexp.MustCompile(`(.*\.pub|authorized_keys|daytona_config|known_hosts|config)$`)

	sshDir := path.Join(os.Getenv("HOME"), "/.ssh")
	files, err := os.ReadDir(sshDir)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if !regex.MatchString(file.Name()) {
				privateKeyOptions = append(privateKeyOptions, huh.NewOption(file.Name(), path.Join(sshDir, file.Name())))
			}
		}
	}

	privateKeyOptions = append(privateKeyOptions, huh.NewOption("Custom path", "custom-path"))

	customPathInput := huh.NewInput().
		Title("SSH private key path").
		Value(privateKeyPath).
		Validate(func(filePath string) error {
			fileInto, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return errors.New("file does not exist")
			} else if err != nil {
				return err
			}

			if fileInto.IsDir() {
				return errors.New("file is a directory")
			}

			return nil
		})

	userInput := huh.NewInput().
		Title("Remote SSH user")

	if user != nil {
		userInput = userInput.Value(user)
	}

	// If there are no private key candidates
	if len(privateKeyOptions) == 1 {
		if user == nil {
			return []*huh.Group{
				huh.NewGroup(
					customPathInput,
				).
					WithHideFunc(func() bool {
						return *authType != SshAuthTypePrivateKey
					}).WithTheme(views.GetCustomTheme()),
			}
		}

		return []*huh.Group{
			huh.NewGroup(
				userInput,
				customPathInput,
			).
				WithHideFunc(func() bool {
					return *authType != SshAuthTypePrivateKey
				}).WithTheme(views.GetCustomTheme()),
		}
	}

	if user == nil {
		return []*huh.Group{
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Remote SSH private key").
					Description("Choose your SSH private key or enter a custom path.").
					Options(
						privateKeyOptions...,
					).
					Value(privateKeyPath),
			).
				WithHideFunc(func() bool {
					return *authType != SshAuthTypePrivateKey
				}).WithTheme(views.GetCustomTheme()),
			huh.NewGroup(
				customPathInput,
			).
				WithHideFunc(func() bool {
					return *privateKeyPath != "custom-path"
				}).WithTheme(views.GetCustomTheme()),
		}
	}

	return []*huh.Group{
		huh.NewGroup(
			userInput,
			huh.NewSelect[string]().
				Title("Remote SSH private key").
				Description("Choose your SSH private key or enter a custom path.").
				Options(
					privateKeyOptions...,
				).
				Value(privateKeyPath),
		).
			WithHideFunc(func() bool {
				return *authType != SshAuthTypePrivateKey
			}).WithTheme(views.GetCustomTheme()),
		huh.NewGroup(
			customPathInput,
		).
			WithHideFunc(func() bool {
				return *privateKeyPath != "custom-path"
			}).WithTheme(views.GetCustomTheme()),
	}
}
