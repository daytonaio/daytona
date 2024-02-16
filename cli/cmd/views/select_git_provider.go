// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views

import (
	"errors"
	"log"

	"github.com/charmbracelet/huh"
)

type GitProviderSelectView struct {
	Id       string
	Username string
	Token    string
}

func GitProviderSelectionView(gitProviderAddView *GitProviderSelectView, isDeleting bool) {

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					huh.NewOption("GitHub", "github"),
					huh.NewOption("GitLab", "gitlab"),
					huh.NewOption("BitBucket", "bitbucket"),
				).
				Value(&gitProviderAddView.Id)),
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(&gitProviderAddView.Username).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("username can not be blank")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return gitProviderAddView.Id != "bitbucket"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Personal access token").
				Value(&gitProviderAddView.Token).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("token can not be blank")
					}
					return nil
				}),
		).WithHide(isDeleting),
	).WithTheme(GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
