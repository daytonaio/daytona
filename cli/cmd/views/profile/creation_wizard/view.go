// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_profile

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/cli/cmd/views"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/config"

	"github.com/charmbracelet/huh"
)

type ProfileAddView struct {
	ProfileName             string
	RemoteHostname          string
	RemoteSshPort           int
	RemoteSshPassword       string
	RemoteSshUser           string
	RemoteSshPrivateKeyPath string
	Provisioner             string
}

func ProfileCreationView(c *config.Config, profileAddView *ProfileAddView, editing bool) {
	var authType views_util.SshAuthType

	privateKeyGroups := views_util.PrivateKeyForm(&authType, &profileAddView.RemoteSshUser, &profileAddView.RemoteSshPrivateKeyPath)

	var portString string = strconv.Itoa(profileAddView.RemoteSshPort)

	formGroups := []*huh.Group{
		huh.NewGroup(
			huh.NewInput().
				Title("Profile name").
				Validate(func(str string) error {
					if str == "" {
						return errors.New("profile name can not be blank")
					}
					if match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", str); !match {
						return errors.New("profile name must be alphanumeric only")
					}

					if !editing {
						for _, profile := range c.Profiles {
							if strings.EqualFold(profile.Name, str) {
								return errors.New("profile name already exists")
							}
						}
					}

					return nil
				}).
				Value(&profileAddView.ProfileName),
			huh.NewInput().
				Title("Remote SSH hostname").
				Value(&profileAddView.RemoteHostname).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("hostname can not be blank")
					}
					return nil
				}),
			huh.NewInput().
				Title("Remote SSH port").
				Value(&portString),
			huh.NewSelect[string]().
				Title("Authentication type").
				Options(
					huh.NewOption("Password", views_util.SshAuthTypePassword),
					huh.NewOption("Private key", views_util.SshAuthTypePrivateKey),
				).
				Value(&authType),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Remote SSH user").
				Value(&profileAddView.RemoteSshUser).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("user can not be blank")
					}
					return nil
				}),
			huh.NewInput().
				Title("Remote SSH password").
				Value(&profileAddView.RemoteSshPassword).
				Password(true).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("password can not be blank")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return authType != "password"
		}),
	}

	formGroups = append(formGroups, privateKeyGroups...)

	form := huh.NewForm(
		formGroups...,
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	profileAddView.RemoteSshPort, err = strconv.Atoi(portString)
	if err != nil {
		log.Fatal(err)
	}
}
