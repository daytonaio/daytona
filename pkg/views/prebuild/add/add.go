// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

var DEFAULT_COMMIT_INTERVAL = "10"
var DEFAULT_RETENTION = "3"

type PrebuildAddView struct {
	WorkspaceConfigName string
	Branch              string
	CommitInterval      string
	TriggerFiles        []string
	Retention           string
	RunBuildOnAdd       bool
}

func PrebuildCreationView(prebuildAddView *PrebuildAddView, editing bool) {
	if !editing && prebuildAddView.CommitInterval == "" {
		prebuildAddView.CommitInterval = DEFAULT_COMMIT_INTERVAL
	}
	if prebuildAddView.Retention == "" {
		prebuildAddView.Retention = DEFAULT_RETENTION
	}

	triggerFilesInput := ""
	for _, triggerFile := range prebuildAddView.TriggerFiles {
		triggerFilesInput += triggerFile + "\n"
	}

	formFields := []huh.Field{
		huh.NewInput().
			Title("Commit interval").
			Description("Leave blank to ignore push events").
			Value(&prebuildAddView.CommitInterval).
			Validate(func(str string) error {
				if str == "" {
					return nil
				}
				num, err := strconv.Atoi(str)
				if err != nil {
					return err
				}
				if num == 0 {
					return errors.New("commit interval cannot be 0")
				}
				return nil
			}),
		huh.NewText().
			Title("Trigger files").
			Description("Enter full paths for files whose changes you want to explicitly trigger a prebuild.\nUse newlines for multiple entries.").
			Value(&triggerFilesInput).Lines(4),
		huh.NewInput().
			Title("Retention").
			Description("Maximum number of resulting builds stored at a time").
			Value(&prebuildAddView.Retention).
			Validate(func(str string) error {
				_, err := strconv.Atoi(str)
				return err
			}),
	}

	if !prebuildAddView.RunBuildOnAdd {
		// Set the default value to true if run flag is not passed
		prebuildAddView.RunBuildOnAdd = true

		formFields = append(formFields, huh.NewConfirm().
			Title("Run the build once on submit?").
			Value(&prebuildAddView.RunBuildOnAdd))
	}

	form := huh.NewForm(
		huh.NewGroup(formFields...),
	).WithTheme(views.GetCustomTheme())

	keyMap := huh.NewDefaultKeyMap()
	keyMap.Text = huh.TextKeyMap{
		NewLine: key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "new line")),
		Next:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		Prev:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
	}

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	prebuildAddView.TriggerFiles = []string{}
	lines := strings.Split(triggerFilesInput, "\n")

	for _, line := range lines {
		if line != "" {
			prebuildAddView.TriggerFiles = append(prebuildAddView.TriggerFiles, strings.TrimRight(line, " "))
		}
	}
}
