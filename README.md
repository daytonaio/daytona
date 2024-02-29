# The Open Source Development Environment Standard
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/569/badge)](https://bestpractices.coreinfrastructure.org/projects/569) [![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/kubernetes)](https://goreportcard.com/report/github.com/kubernetes/kubernetes) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/kubernetes/kubernetes?sort=semver)

<img src="https://github.com/kubernetes/kubernetes/raw/master/logo/logo.png" width="100">

----
Daytona is a radically simple open source development environment standard. It automates the entire process; provisioning the instance, reading the configuration and executing it, setting up prebuilds, attaching your local or a Web IDE, and adding a fully qualified domain name to the dev environment to share and collaborate.

Daytona builds upon a decade and a half of experience of the team responsible for pioneering the Cloud IDE movement. This expertise encompasses the development of IDEs and orchestrators from scratch, facilitating the provision of instant cloud development environments. By integrating modern technologies and advancements with this historical knowledge,  Daytona is committed to eradicating the "works on my machine" syndrome once and for all.

More information on the [Daytona](https://daytona.io/) website.

----

## Features
* __Single Command__: Activate a fully configured development environment with a single command.
* __Runs everywhere__: Spin up your development environment on any machine—whether it's local, remote, a physical server, VM, or cloud-based, & any architecture x86 or Arm.
* __Confiuration File Support__: Initially support for [dev container](https://containers.dev/), ability to expand to DevFile, Nix & Backstage (Contributions welcome here!).
* __Prebuilds System__: Has a prebuilds system, built in, to drasticly improve environment setup times(Contributions welcome here!).
* __IDE Support__ : Seamlessly supports [VS Code](https://github.com/microsoft/vscode) & [JetBrains](https://www.jetbrains.com/remote-development/gateway/) locally, ready to use without configuration. Includes a built-in Web IDE for added convenience.
* __Git Provider Integration__: Start with GitHub support, allowing easy repo or PR selection. Future plans to expand to GitLab & Bitbucket (Contributions welcome here!).
* __Multiple Project Workspace__: Support for multiple project repositories in a dev environment, making it easy to develop using micro-service architecture.
* __Reverse Proxy Integration__: Enable collaboration and streamline feedback loops by leveraging reverse proxy functionality. Access preview environments and the Web IDE seamlessly, even behind firewalls.
* __Extensibility__: Enable extensibility with plugin or provider development. Moreover, in any dynamic language, not just Go(Contributions welcome here!).
* __Security__: Compatibility with all existing technologies development environment technologies
* __Works on my Machine__: Never experince it again.

## Getting Started


### Installing Daytona
Daytona allows you to manage your Development Environments using the Daytona CLI. To install it, please execute the following command:

```bash
curl https://download.daytona.io/daytona/get-server.sh | bash
```

You can also download and compile Daytona from this repository. To do so please check out our docs.

Note the packaged Daytona includes a set of built-in extensions located in the extensions folder, which are added there for your convince.

### Initializing Daytona
To initialize Daytona, follow these steps:

__1. Start the Daytona Service:__  
This initiates the Daytona service, which must always be running for Daytona to function. Use the command:
```bash
daytona server
```
__2. Add Your Git Provider of Choice:__  
Daytona supports GitHub, GitLab, and Bitbucket. To add them to your profile, use the command:  
```bash
daytona git-providers add

```
Follow the steps provided. Here's a link to the [documentation](https://daytona.io/docs) for more details.

__3. Add Your Provider:__  
This step is for choosing where to deploy Development Environments. By default, Daytona includes a Docker provider to spin up environments on your local machine. For remote development environments, use the command:
```bash
daytona providers add
```
This command allows adding connections to an SSH machine or one of the cloud providers (Contributions welcome here!).

__4. Choose Your Default IDE:__  
The default setting for Daytona is VS Code locally. If you prefer, you can switch to VS Code - Browser or any IDE from the JetBrains portfolio (Contributions welcome here!) using the command:
```bash
daytona ide
```





### Setting up your first Daytona Dev environment
```bash
daytona agent install
```

input ssh name (can be localhost)

input username/ password

Note you can use  aws provisioner to auto proivision a vm for each Daytona dev en





## Architecture 
TODO: desc

## How can I extend Daytona
### Provisioners 
### Plugins


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


