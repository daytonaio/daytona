// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jetbrains

func GetIdes() map[Id]Ide {
	return map[Id]Ide{
		CLion:    clion,
		IntelliJ: intellij,
		GoLand:   goland,
		PyCharm:  pycharm,
		PhpStorm: phpstorm,
		WebStorm: webstorm,
		Rider:    rider,
		RubyMine: rubymine,
	}
}

var clion = Ide{
	Name:    "CLion",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/cpp/CLion-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/cpp/CLion-%s-aarch64.tar.gz",
	},
}

var intellij = Ide{
	Name:    "IntelliJ IDEA Ultimate",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/idea/ideaIU-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/idea/ideaIU-%s-aarch64.tar.gz",
	},
}

var goland = Ide{
	Name:    "GoLand",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/go/goland-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/go/goland-%s-aarch64.tar.gz",
	},
}

var pycharm = Ide{
	Name:    "PyCharm Professional",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/python/pycharm-professional-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/python/pycharm-professional-%s-aarch64.tar.gz",
	},
}

var phpstorm = Ide{
	Name:    "PhpStorm",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/webide/PhpStorm-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/webide/PhpStorm-%s-aarch64.tar.gz",
	},
}

var webstorm = Ide{
	Name:    "WebStorm",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/webstorm/WebStorm-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/webstorm/WebStorm-%s-aarch64.tar.gz",
	},
}

var rider = Ide{
	Name:    "Rider",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/rider/JetBrains.Rider-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/rider/JetBrains.Rider-%s-aarch64.tar.gz",
	},
}

var rubymine = Ide{
	Name:    "RubyMine",
	Version: "2023.2.2",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/ruby/RubyMine-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/ruby/RubyMine-%s-aarch64.tar.gz",
	},
}
