// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

var commandAliases = map[string][]string{
	"create":      {"add", "new"},
	"delete":      {"remove", "rm"},
	"update":      {"set"},
	"install":     {"i"},
	"uninstall":   {"u"},
	"info":        {"view", "inspect"},
	"code":        {"open"},
	"logs":        {"log"},
	"forward":     {"fwd"},
	"config":      {"info"},
	"list":        {"ls"},
	"set-default": {"sd"},
	"run":         {"create"},
}

func GetAliases(cmd string) []string {
	if aliases, exists := commandAliases[cmd]; exists {
		return aliases
	}
	return nil
}
