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
		Fleet:    fleet,
	}
}

var clion = Ide{
	ProductCode: "CL",
	Name:        "CLion",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/cpp/CLion-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/cpp/CLion-%s-aarch64.tar.gz",
	},
}

var intellij = Ide{
	ProductCode: "IIU",
	Name:        "IntelliJ IDEA Ultimate",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/idea/ideaIU-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/idea/ideaIU-%s-aarch64.tar.gz",
	},
}

var goland = Ide{
	ProductCode: "GO",
	Name:        "GoLand",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/go/goland-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/go/goland-%s-aarch64.tar.gz",
	},
}

var pycharm = Ide{
	ProductCode: "PCP",
	Name:        "PyCharm Professional",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/python/pycharm-professional-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/python/pycharm-professional-%s-aarch64.tar.gz",
	},
}

var phpstorm = Ide{
	ProductCode: "PS",
	Name:        "PhpStorm",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/webide/PhpStorm-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/webide/PhpStorm-%s-aarch64.tar.gz",
	},
}

var webstorm = Ide{
	ProductCode: "WS",
	Name:        "WebStorm",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/webstorm/WebStorm-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/webstorm/WebStorm-%s-aarch64.tar.gz",
	},
}

var rider = Ide{
	ProductCode: "RD",
	Name:        "Rider",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/rider/JetBrains.Rider-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/rider/JetBrains.Rider-%s-aarch64.tar.gz",
	},
}

var rubymine = Ide{
	ProductCode: "RM",
	Name:        "RubyMine",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/ruby/RubyMine-%s.tar.gz",
		Arm64: "https://download.jetbrains.com/ruby/RubyMine-%s-aarch64.tar.gz",
	},
}

var fleet = Ide{
	ProductCode: "FLL",
	Name:        "Fleet",
	UrlTemplates: UrlTemplates{
		Amd64: "https://download.jetbrains.com/product?code=FLL&release.type=preview&release.type=eap&platform=linux_x64",
		Arm64: "https://download.jetbrains.com/product?code=FLL&release.type=preview&release.type=eap&platform=linux_aarch64",
	},
}
