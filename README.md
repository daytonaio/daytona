
# Instant Local/Remote Development Environments

Daytona Core is a local/remote development environments manager designed for developers. From a single binary, you can create and manage remote development environments, allowing you to focus on your code without the hassle of infrastructure or SaaS.

## Design Principles
* __Single Command__: Activate a fully configured development environment with a single command, eliminating the need for any further user interaction.
* __Host agnostic__: Provision where possible; otherwise, seamlessly connect to existing hosts while abstracting all complexities.
* __Interoperability__:Compatibility with all existing technologies development environment technologies
* __Accessibility__: Enable development environments to be accessible behind firewalls, ensuring utilization and collaboration regardless of network restrictions.
* __Extensibility__: Enable extensibility with plugin development. Moreover, in any dynamic language, not just Go.
* __Works on my Machine__: Never experince it again.



# Table of contents

- [Intro](#intro)
- [Usage](#usage)
- [Plugins](#plugins)
- [Contributing](#contributing)
- [License](#license)

# Intro

[(Back to top)](#table-of-contents)

![Daytona Core](./public/images/daytona.png)

Daytona Core offers individual developers advantages over other solutions for managing development environments. With Daytona Core, you have an easy and open source solution in a single binary. No Kubernetes clusters to setup and manage. No SaaS accounts or limited free hours to work with.

There are two ways to use Daytona Core. If you have got Docker installed (Docker Desktop or Docker Engine) you can use Daytona Core locally to manage your development environments. This is similar to the experience you get from using the dev container CLI. Except that with Daytona Core, you do not have to specify a dev container file. You can use plain git repositories with any container image.

The second way is to use Daytona Core to manage remote development environments. The only requirement is SSH access to a host that can run Docker. If you have a home lab, Linux VM, or access to a developer friendly cloud like Digital Ocean, Civo, Scaleway, etc this is a great option.
## Features2
* __Runs everywhere__: Spin up your development environment on any machine—whether it's local, remote, a physical server, VM, or cloud-based.

* __Confiuration File Support__: Initially support for [dev container](https://containers.dev/), ability to expand to DevFile, Nix & Backstage (Contributions welcome here!).
* __Prebuilds System__: Has a prebuilds system, built in, to drasticly improve environment setup times.
* __IDE Support__ : Seamlessly supports [VS Code](https://github.com/microsoft/vscode) & [JetBrains](https://www.jetbrains.com/remote-development/gateway/) locally, ready to use without configuration. Includes a built-in Web IDE for added convenience.
* __Git Provider Integration__: Start with GitHub support, allowing easy repo or PR selection via dropdown. Future plans to expand to GitLab & Bitbucket (Contributions welcome here!).
* __Reverse Proxy Integration__: Enable collaboration and streamline feedback loops by leveraging reverse proxy functionality. Access preview environments and the Web IDE seamlessly, even behind firewalls.


## Features

* Spin-up ready to code development environments (Workspaces) for all popular programming languages
* Multiple repository Workspace configuration suitable for micro-service architecture
* First class devcontainer.json support
* Easy install on remote machines over SSH
* Prebuild hooks for always ready to code environments
* Plugin system for extending the core feature set
* Available plugins for SSH access, VS Code Server, Tailscale

## Read more about how we think about development

* https://www.daytona.io/dotfiles/embracing-standardized-development-environments
* https://www.daytona.io/dotfiles/mastering-development-environment-configuration-standards
* https://www.daytona.io/dotfiles/the-true-cost-of-developer-tools
* https://www.daytona.io/dotfiles/impact-of-development-environments-on-software-creation

# Usage

[(Back to top)](#table-of-contents)

Using Daytona Core is straightforward. You’ll need a Linux host running Docker. This can be local or remote. Here are the commands you will need to get started:

To setup the Daytona:

Start the server with

```
daytona server
```

In a separate shell or on a different machine you can add a profile for that server.

```
daytona profile add
```

![Create Profile](./public/images/create-profile.png)

To create a new remote development environment, use:

```
daytona create <name> -r <https://github.com/repo/youwant>
```
![Create Workspace](./public/images/create-workspace.png)

You can list the running workspaces with:

```
daytona list
```
![List Workspaces](./public/images/list-workspace.png)

You can open the workspace in VS Code with:

```
daytona open <name>
```

Or you can SSH to the workspace with:

```
daytona ssh <name>
```

![SSH to Workspace](./public/images/ssh-workspace.png)

When you are done:
```
daytona delete <name>
```

You can find more detailed documentation here. <INSERT DOCS LINK>

# Plugins

TODO: desc

* SSH Access
* VS Code Server
* Tailscale

# Building

TODO

# Contributing

[(Back to top)](#table-of-contents)

Daytona is Open Source under the [Apache License 2.0](LICENSE), and is the [copyright of its contributors](NOTICE). If you would like to contribute to the software, you must:

1. Read the [Contributors](CONTRIBUTORS.md) file.
2. Agree to the terms by having a commit in your pull request "signing" the file by adding your name and GitHub handle on a new line at the bottom of the file.
3. Make sure your commits Author metadata matches the name and handle you added to the file.

This ensures that users, distributors, and other contributors can rely on all the software related to Daytona being contributed under the terms of the [License](LICENSE). No contributions will be accepted without following this process.

Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.

# License

[(Back to top)](#table-of-contents)

This repository contains Daytona, covered under the [Apache License 2.0](LICENSE), except where noted (any Daytona logos or trademarks are not covered under the Apache License, and should be explicitly noted by a LICENSE file.)

Daytona is a product produced from this open source software, exclusively by Daytona Platforms, Inc. It is distributed under our commercial terms.

Others are allowed to make their own distribution of the software, but they cannot use any of the Daytona trademarks, cloud services, etc.

We explicitly grant permission for you to make a build that includes our trademarks while developing Daytona itself. You may not publish or share the build, and you may not use that build to run Daytona for any other purpose.

## Code of Conduct
[(Back to top)](#table-of-contents)

This project has adapted the Code of Conduct from the [Contributor Covenant](https://www.contributor-covenant.org/). For more information see the [Code of Conduct](CODE_OF_CONDUCT.md) or contact [codeofconduct@daytona.io.](mailto:codeofconduct@daytona.io) with any additional questions or comments.

## Questions
[(Back to top)](#table-of-contents)

For more information on how to use and develop Daytona, talk to us on
[Slack](https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q) and check out our [documentation](https://www.daytona.io/docs/installation/server/).


