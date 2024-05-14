// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jetbrains

type Ide struct {
	ProductCode  string
	Name         string
	UrlTemplates UrlTemplates
}

type UrlTemplates struct {
	Amd64 string
	Arm64 string
}

type Id string

const (
	CLion    Id = "clion"
	IntelliJ Id = "intellij"
	GoLand   Id = "goland"
	PyCharm  Id = "pycharm"
	PhpStorm Id = "phpstorm"
	WebStorm Id = "webstorm"
	Rider    Id = "rider"
	RubyMine Id = "rubymine"
)
