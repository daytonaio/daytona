// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

func GetProfilePath(profileId string) string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(userConfigDir, "daytona", "profiles", makeFilenameSafe(profileId))
}

func makeFilenameSafe(input string) string {
	base := path.Base(input)
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	safe := reg.ReplaceAllString(base, "_")
	return safe
}

func GenerateIdFromName(name string) string {
	var result strings.Builder

	for _, char := range name {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '-' || char == '_' {
			result.WriteRune(char)
		} else if char == ' ' {
			result.WriteRune('_')
		}
	}

	return strings.ToLower(result.String())
}
