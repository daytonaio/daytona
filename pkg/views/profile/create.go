// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"

	"github.com/charmbracelet/huh"
)

type ProfileAddView struct {
	ProfileName             string
	RemoteHostname          string
	RemoteSshPort           int
	RemoteSshPassword       string
	RemoteSshUser           string
	RemoteSshPrivateKeyPath string
	ApiUrl                  string
}

func ProfileCreationView(c *config.Config, profileAddView *ProfileAddView, editing, forceRemoteAccess, skipName bool) {
	var authType util.SshAuthType

	var portString string = strconv.Itoa(profileAddView.RemoteSshPort)

	hasAccessToRemote := forceRemoteAccess

	if editing {
		hasAccessToRemote = true
	}

	var nameGroup *huh.Group

	nameInput := huh.NewInput().
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
		Value(&profileAddView.ProfileName)

	remoteAccessConfirm := huh.NewConfirm().
		Title("Do you have direct access to the machine running the Daytona Server?").
		Description("If you have direct access, you will be asked to input your SSH credentials.\nIf not, you will only have to provide the API URL of the Daytona Server.").
		Value(&hasAccessToRemote)

	if forceRemoteAccess || editing {
		nameGroup = huh.NewGroup(
			nameInput,
		).WithHideFunc(func() bool {
			return skipName
		})
	} else {
		nameGroup = huh.NewGroup(
			nameInput,
			remoteAccessConfirm,
		).WithHideFunc(func() bool {
			return skipName
		})
	}

	formGroups := []*huh.Group{
		nameGroup,
		huh.NewGroup(
			huh.NewInput().
				Title("Server API URL").
				Value(&profileAddView.ApiUrl),
		).WithHideFunc(func() bool {
			if forceRemoteAccess {
				return true
			}

			if editing {
				return false
			}

			return hasAccessToRemote
		}),
	}

	remoteAuthGroups := remoteAuthGroups(profileAddView, &authType, &portString, &hasAccessToRemote)

	formGroups = append(formGroups, remoteAuthGroups...)

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

func remoteAuthGroups(profileAddView *ProfileAddView, authType *util.SshAuthType, portString *string, hasAccessToRemote *bool) []*huh.Group {
	privateKeyGroups := util.PrivateKeyForm(authType, &profileAddView.RemoteSshUser, &profileAddView.RemoteSshPrivateKeyPath)

	res := []*huh.Group{
		huh.NewGroup(
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
				Value(portString).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("port can not be blank")
					}

					if _, err := strconv.Atoi(str); err != nil {
						return errors.New("port must be a number")
					}

					return nil
				}),
			huh.NewSelect[string]().
				Title("Authentication type").
				Options(
					huh.NewOption("Password", util.SshAuthTypePassword),
					huh.NewOption("Private key", util.SshAuthTypePrivateKey),
				).
				Value(authType),
		).WithHideFunc(func() bool {
			return !*hasAccessToRemote
		}),
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
			return *authType != "password" || !*hasAccessToRemote
		}),
	}

	res = append(res, privateKeyGroups...)

	return res
}
