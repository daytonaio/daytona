# The Open Source Development Environment Standard
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/569/badge)](https://bestpractices.coreinfrastructure.org/projects/569) [![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/kubernetes)](https://goreportcard.com/report/github.com/kubernetes/kubernetes) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/kubernetes/kubernetes?sort=semver)

<img src="https://github.com/kubernetes/kubernetes/raw/master/logo/logo.png" width="100">

----
Daytona is a radically simple open source development environment standard. It automates the entire process; provisioning the instance (if remote), to reading the config and executing it, setting up prebuilds, attaching a local ide, to adding a fully qualified domain name to the dev environment to share and collaborate.

Daytona builds upon a decade and a half of experience of the team responsible for pioneering the Cloud IDE movement, the project Codeanywhere. Using the learning of building IDEs and orchestraters from scratch to be able to provision instant development environments,  incorporating that with technologies that are now available, and building oin the dream tha code works on any maching you run it one the team set out to make that instant dev environments a possibility to all 

In essence Daytona removes all the complexities of setting up dev environments and lets you focus on code. 
More information on the Daytona website.

----

## Design Principles
* __Single Command__: Activate a fully configured development environment with a single command, eliminating the need for any further user interaction.
* __Host agnostic__: Provision where possible; otherwise, seamlessly connect to existing hosts while abstracting all complexities.
* __Interoperability__: Compatibility with all existing technologies development environment technologies
* __Accessibility__: Enable development environments to be accessible behind firewalls, ensuring utilization and collaboration regardless of network restrictions.
* __Extensibility__: Enable extensibility with plugin development. Moreover, in any dynamic language, not just Go.
* __Works on my Machine__: Never experince it again.



## Table of contents

- [Getting Started](##GettingStarted)
  - [CLI](https://github.com/daytonaio/daytona#-CLI)
  - [Dev Environment](https://github.com/daytonaio/daytona#-CLI)
- [Features](##Features)
- [Plugins](##Plugins)
- [Contributing](##Contributing)
- [License](##License)
- [Code of Conduct](##CodeofConductg)
- [Questions](##Questions)


## Getting Started


### Install CLI
To manage Daytona Dev environments you need to have access to the Daytona CLI, you can do this by using brew:

```brew daytona install```

You can also download and compile Daytona from this repository. To do so please check out our docs.

Note the packaged Daytona includes a set of built-in extensions located in the extensions folder, which are added there for your convince.

### Setting up your first Daytona Dev environment
```daytona agent install```

input ssh name (can be localhost)

input username/ password

Note you can use  aws provisioner to auto proivision a vm for each Daytona dev en


## Features


* __Runs everywhere__: Spin up your development environment on any machine—whether it's local, remote, a physical server, VM, or cloud-based.

* __Confiuration File Support__: Initially support for [dev container](https://containers.dev/), ability to expand to DevFile, Nix & Backstage (Contributions welcome here!).
* __Prebuilds System__: Has a prebuilds system, built in, to drasticly improve environment setup times.
* __IDE Support__ : Seamlessly supports [VS Code](https://github.com/microsoft/vscode) & [JetBrains](https://www.jetbrains.com/remote-development/gateway/) locally, ready to use without configuration. Includes a built-in Web IDE for added convenience.
* __Git Provider Integration__: Start with GitHub support, allowing easy repo or PR selection via dropdown. Future plans to expand to GitLab & Bitbucket (Contributions welcome here!).
* __Reverse Proxy Integration__: Enable collaboration and streamline feedback loops by leveraging reverse proxy functionality. Access preview environments and the Web IDE seamlessly, even behind firewalls.
* Spin-up ready to code development environments (Workspaces) for all popular programming languages
* Multiple repository Workspace configuration suitable for micro-service architecture
* First class devcontainer.json support
* Easy install on remote machines over SSH
* Prebuild hooks for always ready to code environments
* Plugin system for extending the core feature set
* Available plugins for SSH access, VS Code Server, Tailscale



## Plugins


TODO: desc

* SSH Access
* VS Code Server
* Tailscale

## Contributing



Daytona is Open Source under the [Apache License 2.0](LICENSE), and is the [copyright of its contributors](NOTICE). If you would like to contribute to the software, you must:

1. Read the [Contributors](CONTRIBUTORS.md) file.
2. Agree to the terms by having a commit in your pull request "signing" the file by adding your name and GitHub handle on a new line at the bottom of the file.
3. Make sure your commits Author metadata matches the name and handle you added to the file.

This ensures that users, distributors, and other contributors can rely on all the software related to Daytona being contributed under the terms of the [License](LICENSE). No contributions will be accepted without following this process.

Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.

## License



This repository contains Daytona, covered under the [Apache License 2.0](LICENSE), except where noted (any Daytona logos or trademarks are not covered under the Apache License, and should be explicitly noted by a LICENSE file.)

Daytona is a product produced from this open source software, exclusively by Daytona Platforms, Inc. It is distributed under our commercial terms.

Others are allowed to make their own distribution of the software, but they cannot use any of the Daytona trademarks, cloud services, etc.

We explicitly grant permission for you to make a build that includes our trademarks while developing Daytona itself. You may not publish or share the build, and you may not use that build to run Daytona for any other purpose.

## Code of Conduct


This project has adapted the Code of Conduct from the [Contributor Covenant](https://www.contributor-covenant.org/). For more information see the [Code of Conduct](CODE_OF_CONDUCT.md) or contact [codeofconduct@daytona.io.](mailto:codeofconduct@daytona.io) with any additional questions or comments.

## Questions


For more information on how to use and develop Daytona, talk to us on
[Slack](https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q) and check out our [documentation](https://www.daytona.io/docs/installation/server/).


